package server

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/mmahmoudian/tmp-place/cmd/shared"
)

type FileInfo struct {
	UploadEpoch      int64
	OriginalFilename string
	TaggedFilename   string
	TTLInSeconds     int64
	DownloadSecret   string
	OneoffDownload   bool
	DeletionEpoch    int64
	FileSHA1         string
}

type FileMetadata struct {
	Size int64
	// SHA1 checksum of the uploaded file
	SHA1Checksum string
}

// GenerateNewTag generates a unique tagged filename for the uploaded file. It checks if the
// tag already exists in the database, and make sure the new tag is unique.
// gets the database file path as input parameter.
// returns the unique TaggedFilename string.
func GenerateNewTag(dbFilePath string) (string, error) {
	// create a new filename with a random tag
	// ensure the generated TaggedFilename is not already in use, otherwise regenerate a new one
	var TaggedFilename string
	for range 100 {
		TaggedFilename = fmt.Sprint(GenerateRandomTag())

		// query database to make sure TaggedFilename is unique
		dbQuery := fmt.Sprintf("SELECT COUNT(*) FROM uploads WHERE tagged_filename='%s' AND deleted=false;", TaggedFilename)
		count, err := shared.QueryOnDatabase(dbFilePath, dbQuery)
		// fmt.Println("Debug: The count object is:", count)
		if err != nil {
			return "", err
		}

		// extract the value of the count and print it
		countValue := count[0]["COUNT(*)"]

		// if count is "0", then the TaggedFilename is unique
		if countValue == 0 {
			return TaggedFilename, nil
		}
	}

	fmt.Println("Alert: Tried 100 times to generate a unique tag, but all existed in database for non-deleted files")

	return "", errors.New("tried 100 times to generate a unique tag, but all existed in database for non-deleted files")
}

// ReceiveFile is a helper function to handle file saving and metadata storage.
// It returns the FileInfo struct, and any error encountered.
func ReceiveFile(file multipart.File, handler *multipart.FileHeader, cfg shared.Config) (FileInfo, error) {
	// get the original filename
	originalFilename := handler.Filename

	// generate a unique tagged filename
	TaggedFilename, err := GenerateNewTag(cfg.Server.Database.DatabaseFile)
	if err != nil {
		// FIXME: the error should be sent to log file and a generic error message returned to user
		return FileInfo{}, err
	}

	// save the file to disk using the tagged filename
	dst, err := os.Create(fmt.Sprintf("%s/%s", cfg.Uploads.Path, TaggedFilename))
	if err != nil {
		return FileInfo{}, err
	}
	defer dst.Close()
	io.Copy(dst, file)

	return FileInfo{
		OriginalFilename: originalFilename,
		TaggedFilename:   TaggedFilename,
		TTLInSeconds:     cfg.Uploads.MaxTTLSeconds,
		DownloadSecret:   "",
		OneoffDownload:   false,
		DeletionEpoch:    0,
	}, nil
}

// GetFileMetadata is a helper function to retrieve file metadata from storage.
// It returns the FileMetadata struct.
func GetFileMetadata(TaggedFilename string, cfg shared.Config) (FileMetadata, error) {
	// get general information of the file
	fileInfo, err := os.Stat(fmt.Sprintf("%s/%s", cfg.Uploads.Path, TaggedFilename))
	if err != nil {
		return FileMetadata{}, err
	}

	// calculate SHA1 checksum of the file
	sha1Sum, err := FileSHA1Sum(fmt.Sprintf("%s/%s", cfg.Uploads.Path, TaggedFilename))
	if err != nil {
		return FileMetadata{}, err
	}

	// return the metadata
	return FileMetadata{
		Size:         fileInfo.Size(),
		SHA1Checksum: sha1Sum,
	}, nil
}

// FileSHA1Sum returns the SHA-1 checksum of the file at the given path
// as a 40-character hex string (like the Unix tool).
func FileSHA1Sum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// create a new SHA-1 hasher
	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
