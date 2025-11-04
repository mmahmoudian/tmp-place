-- Create "uploads" table
CREATE TABLE IF NOT EXISTS uploads (
    "index" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    -- epoch time when the file was uploaded
    upload_epoch INTEGER NOT NULL,
    -- original filename of the uploaded file
    original_filename TEXT NOT NULL,
    -- tagged filename (max 10 alphanumeric characters)
    tagged_filename TEXT CHECK(LENGTH(tagged_filename) <= 10 AND tagged_filename GLOB '[0-9A-Za-z]*') NOT NULL,
    -- file TTL in seconds
    ttl INTEGER UNSIGNED NOT NULL,
    -- indicate if the download is one-off
    oneoff BOOLEAN NOT NULL,
    -- SHA256 checksum of the download secret
    download_secret TEXT CHECK(LENGTH(download_secret) = 64),
    -- Epoch time when the file should be deleted
    deletion_epoch INTEGER UNSIGNED,
    -- Size of the uploaded file in bytes
    file_size INTEGER UNSIGNED NOT NULL,
    -- SHA1 checksum of the uploaded file
    file_checksum TEXT CHECK(LENGTH(file_checksum) = 40) NOT NULL,
    -- indicate if the file has been deleted
    deleted BOOLEAN NOT NULL DEFAULT 0
);

-- Create "log" table
CREATE TABLE IF NOT EXISTS log (
    "index" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    -- epoch time of the event
    epoch INTEGER NOT NULL,
    -- type of the event
    event_type TEXT NOT NULL,
    -- log message
    message TEXT
);

-- Create index on "tagged_filename" for faster lookups
CREATE INDEX IF NOT EXISTS idx_tagged_filename ON uploads(tagged_filename);