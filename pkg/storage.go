package pkg

import (
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

var StoragePrefix string

var loopDelay = time.Hour * 12
var retensionTime = time.Hour * 24 * 12

func retentionLoop() {
	files, err := os.ReadDir(StoragePrefix)
	if err != nil {
		slog.Error("Could not list storage", "err", err)
		goto LOOP
	}
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			slog.Error("Could not load file info", "err", err)
		}
		// if the file is more than 2 weeks old delete it
		if info.ModTime().Add(retensionTime).Before(time.Now()) {
			path := filepath.Join(StoragePrefix, file.Name())
			slog.Info("Deleting file because it exceeds retension threshold", "path", path)
			err = os.RemoveAll(path)
			if err != nil {
				slog.Error("Could not delete file", "err", err)
			}
		}
	}
LOOP:
	time.Sleep(loopDelay)
	retentionLoop()
}

func Initialize(prefix string) {
	StoragePrefix = prefix
	slog.Info("Initialized storage", "path", StoragePrefix)

	err := os.MkdirAll(StoragePrefix, 0700)
	if err != nil {
		log.Fatal("Failed to initialize storage")
	}
	go retentionLoop()
}

type StoreRequest struct {
	Files map[string]io.Reader
}

func Store(request StoreRequest) (string, error) {
	stub := GetName()

	os.MkdirAll(filepath.Join(StoragePrefix, stub), 0700)

	for name, file := range request.Files {
		f, err := os.Create(filepath.Join(StoragePrefix, stub, name))
		if err != nil {
			return "", err
		}
		defer f.Close()

		_, err = io.Copy(f, file)
		if err != nil {
			return "", err
		}
	}

	return stub, nil
}

func List(stub string) ([]string, error) {
	names := []string{}

	files, err := os.ReadDir(filepath.Join(StoragePrefix, stub))
	if err != nil {
		return names, err
	}

	for _, file := range files {
		names = append(names, file.Name())
	}

	return names, nil
}
