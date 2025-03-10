package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type ArchiveFile struct {
	Path string `json:",string"`
	Hash string `json:",string"`
}

func main() {
	fmt.Println("Deduplicating files")
	if len(os.Args) < 2 {
		panic("Please provide at least the folder you want to deduplicate!")
	}

	uniqueFiles, allFiles, duplicates := collectFiles(os.Args[1])
	fmt.Printf("%d unique files and %d files in root path\n", len(uniqueFiles), len(allFiles))

	duplicatesJson, err := json.Marshal(duplicates)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("duplicates.json", duplicatesJson, 0666)
	if err != nil {
		panic(err)
	}
}

func collectFiles(path string) (map[string]string, []ArchiveFile, []ArchiveFile) {
	uniqueFiles := map[string]string{}
	files := []ArchiveFile{}
	duplicates := []ArchiveFile{}

	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			h, err := hashFile(path)
			if err != nil {
				return err
			}

			af := ArchiveFile{
				Path: path,
				Hash: h,
			}
			files = append(files, af)

			_, ok := uniqueFiles[h]
			if ok {
				duplicates = append(duplicates, af)
			} else {
				uniqueFiles[h] = path
			}

		}
		return nil
	})

	return uniqueFiles, files, duplicates
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
