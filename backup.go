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

func Backup(sourceRoot, destinationRoot string, archivedFiles map[string]string) {

	filesToArchive := []ArchiveFile{}

	sourceLen := len(sourceRoot)
	counter := 0

	size, _ := GetSourceSize(sourceRoot)
	fmt.Printf("Source size: %s\n", ByteConverter(size))
	var processedFileSize int64
	processedFileSize = 0
	percent := 0

	filepath.Walk(sourceRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil && !os.IsPermission(err) {
			return err
		}

		if strings.HasPrefix(info.Name(), ".") {
			// skip hidden files
			return nil
		}

		if !info.IsDir() {
			processedFileSize += info.Size()
			p := int((processedFileSize * 100) / size)
			if p > percent {
				percent = p
				fmt.Printf("%d%%\n", percent)
			}

			if info.Size() < SMALL_FILE_SIZE {
				// small file, read into memory
				filecontent, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				h := sha256.New()
				_, err = io.Copy(h, bytes.NewReader(filecontent))
				if err != nil {
					return err
				}

				hash := fmt.Sprintf("%x", h.Sum(nil))
				_, exists := archivedFiles[hash]
				if exists {
					return nil
				}

				newPath := filepath.Join(destinationRoot, path[sourceLen:])
				err = os.MkdirAll(filepath.Dir(newPath), 0755)
				if err != nil {
					return err
				}
				err = os.WriteFile(newPath, filecontent, info.Mode().Perm())
				if err != nil {
					return err
				}
				archivedFiles[hash] = path
				return nil
			}

			h, err := HashFile(path)
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
				counter++
				if counter == 100 {
					fmt.Print("|")
				} else if counter == 200 {
					fmt.Print("/")
				} else if counter == 300 {
					fmt.Print("-")
				} else if counter == 400 {
					fmt.Print("\\")
				} else if counter == 500 {
					fmt.Print("-")
					counter = 0
				}
			}
		}
		return nil
	})

	fileCount := len(filesToArchive)
	fmt.Printf("\n\n%d big files collected", fileCount)
	percent = fileCount / 100
	progress := 0
	progressCounter := 0

	for _, file := range filesToArchive {
		err := CopyFile(sourceRoot, destinationRoot, &file)
		if err != nil {
			fmt.Printf("Error copying file %s: %v\n", file.Path, err)
			continue
		}
		archivedFiles[file.Hash] = file.Path
		progress++
		if progress == percent {
			progress = 0
			progressCounter++
			fmt.Printf("%d%%\n", progressCounter)
		}
	}
}
