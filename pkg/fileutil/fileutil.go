package fileutil

import (
	"fmt"
	"io"
	"os"
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
