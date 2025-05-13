package tasks

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func UnzipAPIM(srcZip string, dest string) (string, error) {
	fmt.Println("🔍 Unzipping APIM...")

	err := os.MkdirAll(dest, 0755)
	if err != nil {
		return "", err
	}

	r, err := zip.OpenReader(srcZip)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, f.Mode())
		} else {
			if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
				return "", err
			}

			inFile, err := f.Open()
			if err != nil {
				return "", err
			}
			defer inFile.Close()

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return "", err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return "", err
			}
		}
	}

	// Assume top-level dir is the one we need
	files, _ := os.ReadDir(dest)
	for _, f := range files {
		if f.IsDir() {
			return filepath.Join(dest, f.Name()), nil
		}
	}

	return "", fmt.Errorf("could not detect unzipped APIM directory")
}