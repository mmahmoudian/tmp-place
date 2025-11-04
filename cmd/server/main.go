package server

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/mmahmoudian/tmp-place/cmd/shared"
	"github.com/spf13/cobra"
)

// ServerRequestHandler handles the file upload and download requests.
func ServerRequestHandler(cfg shared.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, cfg.Uploads.MaxFileSize)
		err := r.ParseMultipartForm(cfg.Uploads.MaxFileSize)
		if err != nil {
			http.Error(w, "File too large", http.StatusRequestEntityTooLarge)
			return
		}

		// get epoch time
		uploadEpoch := time.Now().Unix()

		// Handle TTL
		ttlInSeconds := cfg.Uploads.MaxTTLSeconds
		ttl := r.FormValue("ttl")
		if ttl != "" {
			ttlInSeconds, err = ConvertToSeconds(ttl)
			if err != nil {
				http.Error(w, "Invalid TTL format. Please use a valid duration (e.g., 10s, 5m, 1h)", http.StatusBadRequest)
				return
			}
		}

		// Handle Download Secret
		OriginalDownloadSecret := r.FormValue("secret")
		var DownloadSecret string
		if OriginalDownloadSecret != "" {
			// sanitize and prepare the secret input
			DownloadSecret = PrepareSecret(OriginalDownloadSecret)
		}

		// Handle One-off Download
		oneoff := r.FormValue("oneoff")
		if oneoff != "true" && oneoff != "1" && oneoff != "" {
			http.Error(w, "Invalid One-off Download format. Please use true or 1", http.StatusBadRequest)
			return
		}

		// Handle file argument
		file, handler, err := r.FormFile("file")
		if err == nil {
			defer file.Close()

			// Handle File Upload
			FileInfo, err := ReceiveFile(file, handler, cfg)
			if err != nil {
				http.Error(w, "Unable to save the file", http.StatusInternalServerError)
				return
			}

			// add more information to FileInfo
			FileInfo.UploadEpoch = uploadEpoch
			FileInfo.DownloadSecret = DownloadSecret
			FileInfo.TTLInSeconds = ttlInSeconds
			FileInfo.OneoffDownload = (oneoff == "true" || oneoff == "1")
			FileInfo.DeletionEpoch = uploadEpoch + ttlInSeconds

			// inform the database about the new file
			err = AddFileToDB(FileInfo, cfg)
			if err != nil {
				http.Error(w, "Was unable to save file to database", http.StatusInternalServerError)
				return
			}

			// inform the user about the upload and download details
			fmt.Fprintf(w, "File %s uploaded successfully.\n", FileInfo.OriginalFilename)
			fmt.Fprintf(w, "TTL set to %s.\n", SecondsToHumanReadable(ttlInSeconds))
			fmt.Fprintf(w, "Download secret is: %s\n", OriginalDownloadSecret)
			fmt.Fprintf(w, "One-off download: %t\n", FileInfo.OneoffDownload)
			if DownloadSecret != "" {
				fmt.Fprintf(w, "Use the secret to download your file securely.\n")
				fmt.Fprintf(w, "Download link: %s/%s?secret=%s\n", cfg.Server.Host, FileInfo.TaggedFilename, OriginalDownloadSecret)
			} else {
				fmt.Fprintf(w, "No secret was provided. The file will be publicly accessible.\n")
				fmt.Fprintf(w, "Download link: %s/%s\n", cfg.Server.Host, FileInfo.TaggedFilename)
			}
		} else {
			fmt.Fprintf(w, "No file uploaded.\n")
		}
	}
}

// ServerHandler starts the HTTP server to handle file uploads.
func ServerHandler(cmd *cobra.Command, args []string) {
	// fmt.Println("The server handler is not yet implemented.")

	// Load configuration
	cfg, err := shared.LoadConfig("config.json")
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	// Ensure upload directory exists
	err = os.MkdirAll(cfg.Uploads.Path, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating upload directory:", err)
		os.Exit(1)
	}

	// Set up HTTP server
	http.HandleFunc("/", ServerRequestHandler(cfg))
	fmt.Println("Starting server at :", cfg.Server.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port), nil)
}
