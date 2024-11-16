package io

import (
	"bufio"
	"fmt"
	"os"
)

// ReadStdin reads all of standard input and returns the bytes. If an error
// occurs during scanning, it is returned.
func ReadStdin() ([]byte, error) {
	var b []byte
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		b = append(b, input.Bytes()...)
		b = append(b, byte('\n'))
	}
	if err := input.Err(); err != nil {
		return b, fmt.Errorf("failed to read stdin: %w", err)
	}
	return b, nil
}
