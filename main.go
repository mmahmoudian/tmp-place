package main

import (
    "fmt"
    "io"
    "net/http"
    "os"
)

func handler(w http.ResponseWriter, r *http.Request) {
    // Limit maximum upload size to 10MB
    r.ParseMultipartForm(10 << 20)

    // Get the 'name' field
    name := r.FormValue("name")
    if name == "" {
        name = "stranger"
    }
    fmt.Fprintf(w, "Hello, %s!\n", name)

    // Get the uploaded file
    file, handler, err := r.FormFile("file")
    if err == nil {
        defer file.Close()
        // Save the uploaded file
        dst, err := os.Create(handler.Filename)
        if err != nil {
            http.Error(w, "Unable to save the file", http.StatusInternalServerError)
            return
        }
        defer dst.Close()
        io.Copy(dst, file)
        fmt.Fprintf(w, "File %s uploaded successfully.\n", handler.Filename)
    } else {
        fmt.Fprintf(w, "No file uploaded.\n")
    }
}

func main() {
    http.HandleFunc("/", handler)
    fmt.Println("Starting server at :8080")
    http.ListenAndServe(":8080", nil)
}
