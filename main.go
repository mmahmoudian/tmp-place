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

		name := r.FormValue("name")
		if name == "" {
			name = "stranger"
		}
		fmt.Fprintf(w, "Hello, %s!\n", name)

		file, handler, err := r.FormFile("file")
		if err == nil {
			defer file.Close()
			// get the original filename
			originalFilename := handler.Filename
			// create a new filename with a random tag
			taggedFilename := fmt.Sprintf("%s_%s", helpers.GenerateRandomTag(), originalFilename)
			// save the file to disk using the tagged filename
			dst, err := os.Create(taggedFilename)
			if err != nil {
				http.Error(w, "Unable to save the file", http.StatusInternalServerError)
				return
			}
			defer dst.Close()
			io.Copy(dst, file)
			fmt.Fprintf(w, "File %s uploaded successfully.\n", originalFilename)
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

	// Set up HTTP server
	http.HandleFunc("/", handler(cfg))
	fmt.Println("Starting server at :", cfg.Server.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port), nil)
}
