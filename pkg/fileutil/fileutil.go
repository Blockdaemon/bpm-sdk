package fileutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
)

// Copy a file only if it doesn't exist yet
func CopyFileIfAbsent(src, dst string) error {
	_, err := os.Stat(dst)
	if err == nil {
		fmt.Printf("File %q already exists, skipping copying!\n", dst)
		return nil
	}

	fmt.Printf("Copying %q to %s\n", src, dst)
	return CopyFile(src, dst)
}

// Copy the src file to dst. Any existing file will be overwritten and will not
// copy file attributes.
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func MakeDirectory(baseDir string, subDirs ...string) (string, error) {
	expandedBaseDir, err := homedir.Expand(baseDir)
	if err != nil {
		return "", err
	}

	subDirs = append([]string{expandedBaseDir}, subDirs...)

	path := filepath.Join(subDirs...)

	// Create directory structure if it doesn't exist
	err = os.MkdirAll(path, os.ModePerm)
	return path, err
}

func FileExists(name string) (bool, error) {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}
	return true, nil
}
