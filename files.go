package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Current limit: 1 GiB
// Files under this size will read into RAM.
// In my sample, over 99% of files are under 1 GiB
const SMALL_FILE_SIZE = 1024 * 1024 * 1024

type ArchiveFile struct {
	Path string `json:",string"`
	Hash string `json:",string"`
}

// Returns the size of the source directory in bytes
func GetSourceSize(sourceRoot string) (int64, error) {
	var size int64

	filepath.Walk(sourceRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil && !os.IsPermission(err) {
			fmt.Println(err.Error())
			return nil
		}

		if !info.IsDir() {
			size += info.Size()
		}

		return nil
	})

	return size, nil
}

// Create statiscs file with the size of all files in the source directory
func FieSizeStatiscs(sourceRoot, resultFile string) error {

	fileSizes := []int64{}

	filepath.Walk(sourceRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil && !os.IsPermission(err) {
			fmt.Println(err.Error())
			return nil
		}

		if !info.IsDir() {
			fileSizes = append(fileSizes, info.Size())
		}

		return nil
	})

	file, err := os.Create(resultFile)
	if err != nil {
		fmt.Printf("Error creating %s: %v\n", resultFile, err)
		return err
	}
	defer file.Close()

	for _, size := range fileSizes {
		_, err := fmt.Fprintln(file, size)
		if err != nil {
			break
		}
	}

	return nil
}

// Copy a file, even if it's on a different volume.
func CopyFile(sourceRoot, destinationRoot string, file *ArchiveFile) error {
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

// Collect duplicate files in the source directory
func CollectDuplicates(rootPath string, uniqueFiles map[string]string) ([]ArchiveFile, error) {
	files := []ArchiveFile{}
	duplicates := []ArchiveFile{}

	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			h, err := HashFile(path)
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

	return duplicates, nil
}

func HashFile(path string) (string, error) {
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

func HashBytes(data []byte) (string, error) {
	h := sha256.New()
	_, err := io.Copy(h, bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	hash := fmt.Sprintf("%x", h.Sum(nil))
	return hash, nil
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

// Convert bytes to 2-digit resolution string
func ByteConverter(bytes int64) string {
	if bytes > 1024*1024*1024*1024 {
		return fmt.Sprintf("%.2f TiB", float64(bytes)/1024/1024/1024/1024)
	} else if bytes > 1024*1024*1024 {
		return fmt.Sprintf("%.2f GiB", float64(bytes)/1024/1024/1024)
	} else if bytes > 1024*1024 {
		return fmt.Sprintf("%.2f MiB", float64(bytes)/1024/1024)
	} else if bytes > 1024 {
		return fmt.Sprintf("%.2f KiB", float64(bytes)/1024)
	} else {
		return fmt.Sprintf("%d bytes", bytes)
	}
}
