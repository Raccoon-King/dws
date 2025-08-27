package main

import (
	"fmt"
	"regexp"
)

func main() {
	pattern := "(?s).*test content.*"
	line := "test content"

	re, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Printf("Error compiling regex: %v\n", err)
		return
	}

	fmt.Printf("Compiled Pattern: %s\n", re.String())

	if re.MatchString(line) {
		fmt.Println("Match found!")
	} else {
		fmt.Println("No match.")
	}
}
