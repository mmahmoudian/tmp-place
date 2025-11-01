package server

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/mmahmoudian/tmp-place/cmd/shared"
)

type FileInfo struct {
	OriginalFilename string
	TaggedFilename   string
	TTLInSeconds     int64
	DownloadSecret   string
}

type FileMetadata struct {
	Size int64
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
	}, nil
}

// GetFileMetadata is a helper function to retrieve file metadata from storage.
// It returns the FileMetadata struct.
func GetFileMetadata(TaggedFilename string, cfg shared.Config) (FileMetadata, error) {
	fileInfo, err := os.Stat(fmt.Sprintf("%s/%s", cfg.Uploads.Path, TaggedFilename))
	if err != nil {
		return FileMetadata{}, err
	}

	return FileMetadata{
		Size: fileInfo.Size(),
	}, nil
}
