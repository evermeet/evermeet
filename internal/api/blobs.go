package api

import (
	"fmt"
	"io"
	"net/http"
	"time"

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

	jsonOK(w, map[string]string{
		"hash": hash,
		"url":  "/api/blobs/" + hash,
	})
}

// handleGetBlob serves the raw blob bytes with correct Content-Type.
func (s *Server) handleGetBlob(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")

	rec, err := s.db.GetBlob(r.Context(), hash)
	if err != nil || rec == nil {
		jsonErr(w, http.StatusNotFound, "blob not found")
		return
	}

	f, err := s.blobs.Open(hash)
	if err != nil {
		jsonErr(w, http.StatusNotFound, "blob not found")
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", rec.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeContent(w, r, hash, rec.CreatedAt, f)
}
