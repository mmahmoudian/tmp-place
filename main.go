package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"tmp-place/helpers"
)

func handler(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, cfg.MaxFileSize)
		err := r.ParseMultipartForm(cfg.MaxFileSize)
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
	cfg, err := LoadConfig("config.json")
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}
	http.HandleFunc("/", handler(cfg))
	fmt.Println("Starting server at :8080")
	http.ListenAndServe(":8080", nil)
}
