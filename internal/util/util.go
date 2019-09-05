package util

import (
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
)

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
