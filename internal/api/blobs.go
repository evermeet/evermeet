package api

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/evermeet/evermeet/internal/events"
	"github.com/evermeet/evermeet/internal/store"
	"github.com/go-chi/chi/v5"
)

const (
	maxUploadSize = 10 << 20 // 10 MB
)

var allowedTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

// handleUploadBlob accepts a multipart file upload, stores it content-addressably,
// and records metadata in the database. Returns {"hash": "...", "url": "..."}.
func (s *Server) handleUploadBlob(w http.ResponseWriter, r *http.Request) {
	did := authDID(r)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		jsonErr(w, http.StatusBadRequest, "file too large or invalid multipart")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		jsonErr(w, http.StatusBadRequest, "missing file field")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	// Validate content type
	if _, ok := allowedTypes[contentType]; !ok {
		jsonErr(w, http.StatusBadRequest, fmt.Sprintf("unsupported content type %q; allowed: jpeg, png, gif, webp", contentType))
		return
	}

	data, err := io.ReadAll(io.LimitReader(file, maxUploadSize))
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, "read file")
		return
	}

	hash, err := s.blobs.Put(data)
	if err != nil {
		s.log.Printf("blob put: %v", err)
		jsonErr(w, http.StatusInternalServerError, "store blob")
		return
	}

	if err := s.db.InsertBlob(r.Context(), &store.BlobRecord{
		Hash:        hash,
		ContentType: contentType,
		Size:        int64(len(data)),
		UploadedBy:  did,
		CreatedAt:   time.Now(),
	}); err != nil {
		s.log.Printf("blob db insert: %v", err)
		jsonErr(w, http.StatusInternalServerError, "record blob")
		return
	}
	if err := s.db.InsertBlobSource(r.Context(), &store.BlobSource{
		Hash:        hash,
		InstanceURL: strings.TrimRight(s.baseURL, "/"),
		CreatedAt:   time.Now(),
	}); err != nil {
		s.log.Printf("blob source insert: %v", err)
	}

	jsonOK(w, map[string]string{
		"hash": hash,
		"url":  "/api/blobs/" + hash,
	})
}

// handleGetBlob serves the raw blob bytes with correct Content-Type.
func (s *Server) handleGetBlob(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	if !validBlobHash(hash) {
		jsonErr(w, http.StatusBadRequest, "invalid blob hash")
		return
	}

	rec, err := s.db.GetBlob(r.Context(), hash)
	if err != nil || rec == nil {
		rec, err = s.fetchBlobFromSources(r.Context(), hash)
		if err != nil || rec == nil {
			jsonErr(w, http.StatusNotFound, "blob not found")
			return
		}
	}

	f, err := s.blobs.Open(hash)
	if err != nil {
		rec, err = s.fetchBlobFromSources(r.Context(), hash)
		if err != nil || rec == nil {
			jsonErr(w, http.StatusNotFound, "blob not found")
			return
		}
		f, err = s.blobs.Open(hash)
		if err != nil {
			jsonErr(w, http.StatusNotFound, "blob not found")
			return
		}
	}
	defer f.Close()

	w.Header().Set("Content-Type", rec.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeContent(w, r, hash, rec.CreatedAt, f)
}

func (s *Server) fetchBlobFromSources(ctx context.Context, hash string) (*store.BlobRecord, error) {
	sources, err := s.db.ListBlobSources(ctx, hash)
	if err != nil {
		return nil, err
	}
	localBase := strings.TrimRight(s.baseURL, "/")
	for _, source := range sources {
		instanceURL := strings.TrimRight(source.InstanceURL, "/")
		if instanceURL == "" || instanceURL == localBase {
			continue
		}
		rec, err := s.fetchBlobFromSource(ctx, hash, instanceURL)
		if err != nil {
			s.log.Printf("blob fetch %s from %s failed: %v", hash, instanceURL, err)
			continue
		}
		return rec, nil
	}
	return nil, fmt.Errorf("blob source unavailable")
}

func (s *Server) fetchBlobFromSource(ctx context.Context, hash, instanceURL string) (*store.BlobRecord, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, instanceURL+"/api/blobs/"+url.PathEscape(hash), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("remote returned %s", resp.Status)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxUploadSize+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxUploadSize {
		return nil, fmt.Errorf("blob too large")
	}
	storedHash, err := s.blobs.Put(data)
	if err != nil {
		return nil, err
	}
	if storedHash != hash {
		return nil, fmt.Errorf("blob hash mismatch")
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	contentType = strings.Split(contentType, ";")[0]
	if _, ok := allowedTypes[contentType]; !ok {
		return nil, fmt.Errorf("unsupported content type %q", contentType)
	}

	rec := &store.BlobRecord{
		Hash:        hash,
		ContentType: contentType,
		Size:        int64(len(data)),
		UploadedBy:  "remote",
		CreatedAt:   time.Now(),
	}
	if err := s.db.InsertBlob(ctx, rec); err != nil {
		return nil, err
	}
	if err := s.db.InsertBlobSource(ctx, &store.BlobSource{
		Hash:        hash,
		InstanceURL: instanceURL,
		CreatedAt:   time.Now(),
	}); err != nil {
		s.log.Printf("blob source insert: %v", err)
	}
	return rec, nil
}

func (s *Server) recordEventBlobSources(ctx context.Context, hostURL string, ms events.MutableState) {
	hostURL = strings.TrimRight(hostURL, "/")
	if hostURL == "" || hostURL == strings.TrimRight(s.baseURL, "/") {
		return
	}
	payload, err := json.Marshal(ms)
	if err != nil {
		return
	}
	for hash := range blobHashesInJSON(payload) {
		if err := s.db.InsertBlobSource(ctx, &store.BlobSource{
			Hash:        hash,
			InstanceURL: hostURL,
			CreatedAt:   time.Now(),
		}); err != nil {
			s.log.Printf("blob source insert: %v", err)
		}
	}
}

func blobHashesInJSON(payload []byte) map[string]struct{} {
	var v any
	if err := json.Unmarshal(payload, &v); err != nil {
		return nil
	}
	out := map[string]struct{}{}
	collectBlobHashes(v, out)
	return out
}

func collectBlobHashes(v any, out map[string]struct{}) {
	switch x := v.(type) {
	case string:
		if hash := blobHashFromURL(x); hash != "" {
			out[hash] = struct{}{}
		}
	case []any:
		for _, item := range x {
			collectBlobHashes(item, out)
		}
	case map[string]any:
		for _, item := range x {
			collectBlobHashes(item, out)
		}
	}
}

func blobHashFromURL(value string) string {
	if strings.HasPrefix(value, "/api/blobs/") {
		hash := strings.TrimPrefix(value, "/api/blobs/")
		if i := strings.IndexAny(hash, "?#"); i >= 0 {
			hash = hash[:i]
		}
		if validBlobHash(hash) {
			return hash
		}
	}
	if u, err := url.Parse(value); err == nil && u.Path != "" {
		if strings.HasPrefix(u.Path, "/api/blobs/") {
			hash := strings.TrimPrefix(u.Path, "/api/blobs/")
			if validBlobHash(hash) {
				return hash
			}
		}
	}
	return ""
}

func validBlobHash(hash string) bool {
	if len(hash) != 64 {
		return false
	}
	_, err := hex.DecodeString(hash)
	return err == nil
}
