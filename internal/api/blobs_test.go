package api

import "testing"

func TestBlobHashesInJSON(t *testing.T) {
	hash := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	payload := []byte(`{
		"cover_url": "/api/blobs/` + hash + `",
		"nested": {
			"avatar": "http://localhost:8080/api/blobs/` + hash + `?v=1",
			"ignored": "/api/blobs/not-a-hash"
		}
	}`)

	got := blobHashesInJSON(payload)
	if _, ok := got[hash]; !ok {
		t.Fatalf("expected hash %s in %+v", hash, got)
	}
	if len(got) != 1 {
		t.Fatalf("expected one unique hash, got %+v", got)
	}
}
