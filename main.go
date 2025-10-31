package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"tmp-place/helpers"
)

func handler(cfg helpers.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, cfg.Uploads.MaxFileSize)
		err := r.ParseMultipartForm(cfg.Uploads.MaxFileSize)
		if err != nil {
			http.Error(w, "File too large", http.StatusRequestEntityTooLarge)
			return
		}

		// Handle TTL
		ttlInSeconds := cfg.Uploads.MaxTTLSeconds
		ttl := r.FormValue("ttl")
		if ttl != "" {
			ttlInSeconds, err = helpers.ConvertToSeconds(ttl)
			if err != nil {
				http.Error(w, "Invalid TTL format. Please use a valid duration (e.g., 10s, 5m, 1h)", http.StatusBadRequest)
				return
			}
		}

		// Handle Download Secret
		DownloadSecret := r.FormValue("secret")

		// Handle File Upload
		file, handler, err := r.FormFile("file")
		if err == nil {
			defer file.Close()
			// get the original filename
			originalFilename := handler.Filename
			// create a new filename with a random tag
			taggedFilename := fmt.Sprint(helpers.GenerateRandomTag())
			// save the file to disk using the tagged filename

			dst, err := os.Create(fmt.Sprintf("%s/%s", cfg.Uploads.Path, taggedFilename))
			if err != nil {
				http.Error(w, "Unable to save the file", http.StatusInternalServerError)
				return
			}
			defer dst.Close()
			io.Copy(dst, file)

			// inform the user about the upload and download details
			fmt.Fprintf(w, "File %s uploaded successfully.\n", originalFilename)
			fmt.Fprintf(w, "TTL set to %s.\n", helpers.SecondsToHumanReadable(ttlInSeconds))
			fmt.Fprintf(w, "Download secret is: %s\n", DownloadSecret)
			if DownloadSecret != "" {
				fmt.Fprintf(w, "Use the secret to download your file securely.\n")
				fmt.Fprintf(w, "Download link: %s/%s?secret=%s\n", cfg.Server.Host, taggedFilename, DownloadSecret)
			} else {
				fmt.Fprintf(w, "No secret was provided. The file will be publicly accessible.\n")
				fmt.Fprintf(w, "Download link: %s/%s\n", cfg.Server.Host, taggedFilename)
			}
		} else {
			fmt.Fprintf(w, "No file uploaded.\n")
		}
	}
}

func main() {
	// Load configuration
	cfg, err := helpers.LoadConfig("config.json")
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
	http.HandleFunc("/", handler(cfg))
	fmt.Println("Starting server at :", cfg.Server.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port), nil)
}
