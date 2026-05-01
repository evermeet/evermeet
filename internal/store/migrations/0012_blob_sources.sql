CREATE TABLE blob_sources (
    hash         TEXT NOT NULL,
    instance_url TEXT NOT NULL,
    created_at   TEXT NOT NULL,
    PRIMARY KEY (hash, instance_url)
);
CREATE INDEX idx_blob_sources_hash ON blob_sources(hash);
