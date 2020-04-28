package util

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

func Unzip(unzipFileFullPath string, targetDir string) (string, error) {
	zipFile, err := zip.OpenReader(unzipFileFullPath)
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	if err = os.MkdirAll(targetDir, 0755); err != nil {
		return "", err
	}
	for _, file := range zipFile.File {
		path := filepath.Join(targetDir, file.Name)
		if file.FileInfo().IsDir() {
			if err = os.Mkdir(path, file.Mode()); err != nil {
				return "", err
			}
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return "", err
		}

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return "", err
		}
		fileReader.Close()
		targetFile.Close()
	}
	return "", nil
}


