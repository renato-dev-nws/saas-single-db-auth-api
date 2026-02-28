package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash := "$2a$12$ns1YP4G3P8iRUKwREqMK8eGgIcxvPyAzXxmNibXydt5GRD6LslLG."
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("admin123"))
	fmt.Printf("admin123 match: %v\n", err == nil)
	if err != nil {
		fmt.Println("Error:", err)
	}

	// Generate new hash
	newHash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), 12)
	fmt.Printf("New hash: %s\n", newHash)
}
