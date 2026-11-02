package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <hostname>\nПример: geotracer google.com\n", os.Args[0])

		os.Exit(1)
	}

	target := args[1]

	fmt.Printf("Целевой хост: %s\n", target)
	fmt.Println("Начинаем трассировку...")
}
