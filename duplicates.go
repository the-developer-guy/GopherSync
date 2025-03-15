package main

import (
	"fmt"
)

func Deduplicate(sourceRoot, duplicateBackupRoot string) (map[string]string, error) {

	uniqueFiles := map[string]string{}
	duplicates, err := CollectDuplicates(sourceRoot, uniqueFiles)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%d unique files and %d duplicates in root path\n", len(uniqueFiles), len(duplicates))

	err = MoveFiles(sourceRoot, duplicateBackupRoot, duplicates)
	if err != nil {
		return nil, err
	}

	return uniqueFiles, nil
}
