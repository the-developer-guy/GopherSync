package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	fmt.Println("Deduplicating files")
	if len(os.Args) < 2 {
		panic("Please provide at least the folder you want to deduplicate!")
	}

	collectFiles(os.Args[1])
}

func collectFiles(path string) map[string]string {
	files := map[string]string{}

	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			h, err := hashFile(path)
			if err != nil {
				return err
			}
			files[h] = path
		}
		return nil
	})

	return files
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return "", nil
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
