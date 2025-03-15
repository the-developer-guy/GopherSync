package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Back up the source directory to the destination directory
// using the archivedFiles map to avoid copying files that have
// already been copied.
func Backup(sourceRoot, destinationRoot string,
	archivedFiles map[string]string) {

	filesToArchive := []ArchiveFile{}
	sourceLen := len(sourceRoot)

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
			if info.Size() < SMALL_FILE_SIZE {

				processedFileSize += info.Size()
				p := int((processedFileSize * 100) / size)
				if p > percent {
					percent = p
					fmt.Printf("%d%%\n", percent)
				}

				// small file, read into memory
				filecontent, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				h, _ := HashBytes(filecontent)
				_, exists := archivedFiles[h]
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
				archivedFiles[h] = path
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
			}
		}
		return nil
	})

	bigFileCount := len(filesToArchive)
	fmt.Printf("\n\n%d big files collected", bigFileCount)
	processedBigFileCount := 0

	for _, file := range filesToArchive {
		err := CopyFile(sourceRoot, destinationRoot, &file)
		if err != nil {
			fmt.Printf("Error copying file %s: %v\n", file.Path, err)
			continue
		}
		archivedFiles[file.Hash] = file.Path

		processedBigFileCount++
		fmt.Printf("%.1f%%\n", float64(processedBigFileCount*100)/float64(bigFileCount))
	}
}
