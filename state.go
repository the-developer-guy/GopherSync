package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func LoadStatefile(path string) (map[string]string, error) {
	archivedFiles := map[string]string{}

	statefile, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return archivedFiles, err
	}

	err = json.Unmarshal(statefile, &archivedFiles)
	if err != nil {
		fmt.Println(err)
		return archivedFiles, err
	}

	return archivedFiles, nil
}

func StoreStatefile(path string, archivedFiles map[string]string) error {
	statefileJson, err := json.Marshal(archivedFiles)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = os.WriteFile(path, statefileJson, 0666)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
