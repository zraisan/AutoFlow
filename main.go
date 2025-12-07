package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	fmt.Println("Enter Project Path:")
	var dir string
	n, err := fmt.Scanln(&dir)
	if n == 0 {
		dir = "./"
	}
	fmt.Println(n, dir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range entries {
		fmt.Println(entry.Name())
	}
}
