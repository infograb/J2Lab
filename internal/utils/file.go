package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// FileExists checks if a file exists at the given path
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// CopyFile copies a file from srcPath to destPath
func CopyFile(w io.Writer, srcPath, destPath string) error {
	// Check if destination file exists
	if FileExists(destPath) {
		fmt.Fprintf(w, "File %s already exists. Do you want to override? (y/n): ", destPath)
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		if input != "y\n" {
			return nil
		}
	}

	// Open the source file for reading
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("error opening source file: %v", err)
	}
	defer srcFile.Close()

	// Create or override the destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error creating destination file: %v", err)
	}
	defer destFile.Close()

	// Copy the contents from the source file to the destination file
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("error copying file: %v", err)
	}

	return nil
}
