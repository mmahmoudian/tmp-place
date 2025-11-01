-- Create "uploads" table
CREATE TABLE IF NOT EXISTS uploads (
    "index" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    upload_epoch INTEGER NOT NULL,
    original_filename TEXT NOT NULL,
    tagged_filename TEXT CHECK(LENGTH(tagged_filename) <= 10 AND tagged_filename GLOB '[0-9A-Za-z]*') NOT NULL,
    ttl INTEGER UNSIGNED NOT NULL,
    oneoff BOOLEAN NOT NULL,
    download_secret TEXT CHECK(LENGTH(download_secret) = 40) NOT NULL, -- SHA1 length
    deletion_epoch INTEGER UNSIGNED
);

-- Create "log" table
CREATE TABLE IF NOT EXISTS log (
    "index" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    upload_epoch INTEGER NOT NULL,
    type TEXT NOT NULL,
    message TEXT
);

-- Create index on "tagged_filename" for faster lookups
CREATE INDEX IF NOT EXISTS idx_tagged_filename ON uploads(tagged_filename);