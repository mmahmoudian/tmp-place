package server

import (
	"crypto/sha1"
	"encoding/hex"
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

// ReceiveFile is a helper function to handle file saving and metadata storage.
// It returns the FileInfo struct, and any error encountered.
func ReceiveFile(file multipart.File, handler *multipart.FileHeader, cfg shared.Config) (FileInfo, error) {
	// get the original filename
	originalFilename := handler.Filename

	// create a new filename with a random tag
	// TODO: ensure the generated TaggedFilename is not already in use, otherwise regenerate a new one
	TaggedFilename := fmt.Sprint(GenerateRandomTag())

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
