-- Content-addressable blob metadata. The actual file lives on disk.
CREATE TABLE blobs (
    hash        TEXT PRIMARY KEY,       -- blake3 hex digest (also the filename on disk)
    content_type TEXT NOT NULL,
    size        INTEGER NOT NULL,
    uploaded_by TEXT NOT NULL,          -- uploader DID
    created_at  TEXT NOT NULL
);
