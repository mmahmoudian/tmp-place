package server

import (
	"fmt"

	"github.com/mmahmoudian/tmp-place/cmd/shared"
)

// AddFileToDB is a helper function to store file data in the database.
// It returns any error encountered.
func AddFileToDB(FileInfo FileInfo, cfg shared.Config) error {
	FileMetadata, err := GetFileMetadata(FileInfo.TaggedFilename, cfg)
	if err != nil {
		return err
	}

	// Here you would add code to insert FileInfo and FileMetadata into the database
	// For example, using an SQL INSERT statement
	queryString := `INSERT INTO uploads (
		upload_epoch,
		original_filename,
		tagged_filename,
		ttl,
		oneoff,
		download_secret,
		deletion_epoch,
		file_size,
		file_checksum,
		deleted
		   ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = shared.QueryOnDatabase(
		cfg.Server.Database.DatabaseFile, // database file name
		queryString,                      // the query string
		FileInfo.UploadEpoch,             // upload_epoch
		FileInfo.OriginalFilename,        // original_filename
		FileInfo.TaggedFilename,          // tagged_filename
		FileInfo.TTLInSeconds,            // ttl
		FileInfo.OneoffDownload,          // oneoff
		FileInfo.DownloadSecret,          // download_secret
		FileInfo.DeletionEpoch,           // deletion_epoch
		FileMetadata.Size,                // file_size
		FileMetadata.SHA1Checksum,        // file_checksum
		false,                            // deleted
	)

	if err != nil {
		fmt.Println("Error:", err)
	}

	return err
}
