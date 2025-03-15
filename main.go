package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: gosync sourceRoot duplicateBackupRoot statefilePath")
		os.Exit(1)
	}

	uniqueFiles, err := LoadStatefile(os.Args[3])
	if err != nil {
		fmt.Println(err)
	}
	duplicates, err := CollectDuplicates(os.Args[1], uniqueFiles)
	if err != nil {
		fmt.Println(err)
	}

	StoreStatefile(os.Args[3], uniqueFiles)

	err = MoveFiles(os.Args[1], os.Args[2], duplicates)
	if err != nil {
		fmt.Println(err)
	}
}
