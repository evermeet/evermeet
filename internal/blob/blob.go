// Package blob provides content-addressable blob storage on disk.
// Files are named by their blake3 hash encoded as base32-hex (bafk... style).
package blob

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"lukechampine.com/blake3"
)

// Store writes and reads blobs from a directory on disk.
// Filenames are the content hash, so writes are idempotent.
type Store struct {
	dir string
}

// New creates a Store rooted at dir, creating it if needed.
func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("blob store: mkdir %s: %w", dir, err)
	}
	return &Store{dir: dir}, nil
}

// Put writes data to disk and returns the content hash (hex-encoded blake3).
// If the blob already exists the write is skipped.
func (s *Store) Put(data []byte) (string, error) {
	hash := hashBytes(data)
	path := s.path(hash)

	if _, err := os.Stat(path); err == nil {
		return hash, nil // already stored
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return "", fmt.Errorf("blob write tmp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return "", fmt.Errorf("blob rename: %w", err)
	}
	return hash, nil
}

// Get returns the blob data for the given hash.
func (s *Store) Get(hash string) ([]byte, error) {
	data, err := os.ReadFile(s.path(hash))
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("blob not found: %s", hash)
	}
	return data, err
}

// Open returns a ReadSeeker for streaming large blobs.
func (s *Store) Open(hash string) (io.ReadSeekCloser, error) {
	f, err := os.Open(s.path(hash))
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("blob not found: %s", hash)
	}
	return f, err
}

func (s *Store) path(hash string) string {
	return filepath.Join(s.dir, hash)
}

func hashBytes(data []byte) string {
	h := blake3.Sum256(data)
	return hex.EncodeToString(h[:])
}
