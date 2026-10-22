CREATE TABLE files (
    id            uuid PRIMARY KEY DEFAULT uuidv7(),
    storage_key   text NOT NULL UNIQUE CHECK (storage_key <> ''),
    original_name text NOT NULL CHECK (original_name <> ''),
    mime_type     text NOT NULL CHECK (mime_type <> ''),
    size_bytes    bigint NOT NULL CHECK (size_bytes >= 0),
    sha256        bytea NOT NULL CHECK (octet_length(sha256) = 32),
    visibility    text NOT NULL DEFAULT 'private'
                  CHECK (visibility IN ('public', 'private')),
    width         integer CHECK (width > 0),
    height        integer CHECK (height > 0),
    created_at    timestamptz NOT NULL DEFAULT now(),
    deleted_at    timestamptz,
    CHECK ((width IS NULL) = (height IS NULL))
);

CREATE INDEX files_live_created_at_idx
    ON files (created_at DESC)
    WHERE deleted_at IS NULL;

ALTER TABLE beans
    ADD COLUMN image_file_id uuid REFERENCES files(id) ON DELETE SET NULL;

ALTER TABLE beans DROP COLUMN image_link;
