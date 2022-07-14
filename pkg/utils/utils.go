package utils

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
)

func GOOSPath(path string) string {
	if runtime.GOOS == "windows" {
		return "file:////%3F/" + filepath.ToSlash(path)
	}

	return path
}

func IsDirEmpty(name string) bool {
	f, err := os.Open(name)
	if err != nil {
		return os.IsNotExist(err)
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	return err == io.EOF
}

func FileExist(file string) bool {
	_, err := os.Stat(file)
	return !os.IsNotExist(err)
}
