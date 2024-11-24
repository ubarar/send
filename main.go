package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ubarar/send/pkg"
)

const MaxMemory = 1000000000 // 1GB

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./upload.html")
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./view.html")
}

func uploadMultipleHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(MaxMemory)
	if err != nil {
		slog.Error("Could not parse upload", "err", err)
		http.Error(w, "upload failed", http.StatusBadRequest)
		return
	}

	headers := r.MultipartForm.File["myFiles"]

	// storage request maps names to io.Readers
	request := pkg.StoreRequest{Files: map[string]io.Reader{}}
	for _, header := range headers {
		file, err := header.Open()
		if err != nil {
			slog.Error("Could not open uploaded file", "err", err)
			http.Error(w, "upload failed", http.StatusBadRequest)
			return
		}
		defer file.Close()
		request.Files[header.Filename] = file
	}

	// store the files
	stub, err := pkg.Store(request)
	if err != nil {
		slog.Error("Failed to store files", "err", err)
		http.Error(w, "upload failed", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/view/"+stub, http.StatusSeeOther)
}

func listFilesHandler(w http.ResponseWriter, r *http.Request) {
	stub := strings.TrimPrefix(r.URL.Path, "/list/")
	files, err := pkg.List(stub)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	response, _ := json.Marshal(files)
	fmt.Fprintf(w, string(response))
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	// make an assumption about where the file is
	http.ServeFile(w, r, filepath.Join(pkg.StoragePrefix, strings.TrimPrefix(r.URL.Path, "/download/")))
}

func main() {
	storage := flag.String("storage", "./storage", "where files will be stored")
	flag.Parse()

	pkg.Initialize(*storage)

	// Serves the main page where you can upload a file
	http.HandleFunc("/", http.RedirectHandler("/upload", http.StatusSeeOther).ServeHTTP)

	http.HandleFunc("/upload", uploadHandler)
	// When you are given a valid link, this page shows you all the files
	// that were uploaded
	http.HandleFunc("/view/", viewHandler)

	// Upload multiple files to this URL and in return you get a code
	// that will create the link to view the files again
	http.HandleFunc("/uploadmultiple", uploadMultipleHandler)

	http.HandleFunc("/list/", listFilesHandler)
	// Use this handler to download the file
	http.HandleFunc("/download/", downloadHandler)

	// serve up css and all
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	slog.Error("Terminating", http.ListenAndServe(":8080", nil))
}
