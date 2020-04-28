package handler

import (
	"fmt"
	"io"
	"os"
)

func MoveFile(oldPath string, newPath string) error {
	oldFile, err := os.Open(oldPath)
	if err != nil {
		return fmt.Errorf("Couldn't open source file: %s ", err)
	}
	newFile, err := os.Create(newPath)
	if err != nil {
		oldFile.Close()
		return fmt.Errorf("Couldn't open dest file: %s ", err)
	}
	defer newFile.Close()
	_, err = io.Copy(newFile, oldFile)
	newFile.Close()
	if err != nil {
		return fmt.Errorf("Writing to output file failed: %s ", err)
	}
	err = os.Remove(oldPath)
	if err != nil {
		return fmt.Errorf("Failed removing original file: %s ", err)
	}
	return nil
}
