package main

import (
	"encoding/json"
	"os"
)

func LoadStatefile(path string) (map[string]string, error) {
	statefile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	archivedFiles := map[string]string{}
	err = json.Unmarshal(statefile, &archivedFiles)
	if err != nil {
		return nil, err
	}

	return archivedFiles, nil
}

func StoreStatefile(path string, archivedFiles map[string]string) error {
	statefileJson, err := json.Marshal(archivedFiles)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, statefileJson, 0666)
	if err != nil {
		return err
	}

	return nil
}
