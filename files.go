package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ArchiveFile struct {
	Path string `json:",string"`
	Hash string `json:",string"`
}

func Backup(sourceRoot, destinationRoot string, archivedFiles map[string]string) {

	filesToArchive := []ArchiveFile{}

	filepath.Walk(sourceRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			h, err := hashFile(path)
			if err != nil {
				return err
			}

			_, exists := archivedFiles[h]
			if !exists {
				af := ArchiveFile{
					Path: path,
					Hash: h,
				}
				filesToArchive = append(filesToArchive, af)
			}
		}
		return nil
	})

	for _, file := range filesToArchive {
		err := copyFile(sourceRoot, destinationRoot, &file)
		if err != nil {
			fmt.Printf("Error copying file %s: %v\n", file.Path, err)
			continue
		}
		archivedFiles[file.Hash] = file.Path
	}
}

func copyFile(sourceRoot, destinationRoot string, file *ArchiveFile) error {
	sourceLen := len(sourceRoot)

	if !strings.HasPrefix(file.Path, sourceRoot) {
		return fmt.Errorf("file %s is not in the expected %s path", file.Path, sourceRoot)
	}

	newPath := filepath.Join(destinationRoot, file.Path[sourceLen:])
	err := os.MkdirAll(filepath.Dir(newPath), 0755)
	if err != nil {
		return err
	}

	sourceFile, err := os.Open(file.Path)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(newPath)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}

func Deduplicate() {

	uniqueFiles, allFiles, duplicates := collectFiles(os.Args[1])
	fmt.Printf("%d unique files and %d files in root path\n", len(uniqueFiles), len(allFiles))

	statefileJson, err := json.Marshal(uniqueFiles)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("/Volumes/Master/statefile.json", statefileJson, 0666)
	if err != nil {
		panic(err)
	}

	duplicatesJson, err := json.Marshal(duplicates)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("duplicates.json", duplicatesJson, 0666)
	if err != nil {
		panic(err)
	}

	if len(os.Args) == 3 {
		fmt.Printf("Moving duplicate files from %s to %s\n", os.Args[1], os.Args[2])
		MoveFiles(os.Args[1], os.Args[2], duplicates)
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

func DeleteFiles(files []ArchiveFile) error {
	for _, file := range files {
		err := os.Remove(file.Path)
		if err != nil {
			return err
		}
	}

	return nil
}

func MoveFiles(sourceRoot, destinationRoot string, files []ArchiveFile) error {
	sourceLen := len(sourceRoot)

	for _, file := range files {
		if !strings.HasPrefix(file.Path, sourceRoot) {
			return fmt.Errorf("file %s is not in the expected %s path", file.Path, sourceRoot)
		}

		newPath := filepath.Join(destinationRoot, file.Path[sourceLen:])
		err := os.MkdirAll(filepath.Dir(newPath), 0755)
		if err != nil {
			return err
		}
		err = os.Rename(file.Path, newPath)
		if err != nil {
			return err
		}
	}

	return nil
}
