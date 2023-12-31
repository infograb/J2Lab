/*
 * This file is part of the InfoGrab project.
 *
 * Copyright (C) 2023 InfoGrab
 *
 * This program is free software: you can redistribute it and/or modify it
 * it is available under the terms of the GNU Lesser General Public License
 * by the Free Software Foundation, either version 3 of the License or by the Free Software Foundation
 * (at your option) any later version.
 */

package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
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
		return errors.Wrap(err, "error opening source file: %v")
	}
	defer srcFile.Close()

	// Create or override the destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return errors.Wrap(err, "error creating destination file: %v")
	}
	defer destFile.Close()

	// Copy the contents from the source file to the destination file
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return errors.Wrap(err, "error copying file: %v")
	}

	return nil
}
