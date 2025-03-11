package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Deduplicating files")
	if len(os.Args) < 2 {
		panic("Please provide at least the folder you want to deduplicate!")
	}

}
