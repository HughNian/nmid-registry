package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var URLFriendlyCharactersRegex = regexp.MustCompile(`^[A-Za-z0-9\-_\.~]{1,253}$`)

func GOOSPath(path string) string {
	if runtime.GOOS == "windows" {
		return "file:////%3F/" + filepath.ToSlash(path)
	}

	return path
}

func Exit(code int, msg string) {
	if code != 0 {
		if msg != "" {
			fmt.Fprintf(os.Stderr, "%s\n", msg)
		}
		os.Exit(code)
	}

	if msg != "" {
		fmt.Fprintf(os.Stdout, "%v\n", msg)
	}

	os.Exit(0)
}

func ValidateName(name string) error {
	if !URLFriendlyCharactersRegex.Match([]byte(name)) {
		return fmt.Errorf("invalid constant name %s", name)
	}

	return nil
}

func GetMemberName(apiAddr string) (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	memberName := hostname + "-" + apiAddr
	memberName = strings.Replace(memberName, ",", "-", -1)
	memberName = strings.Replace(memberName, ":", "-", -1)
	memberName = strings.Replace(memberName, "=", "-", -1)

	return memberName, nil
}

func FileExist(file string) bool {
	_, err := os.Stat(file)
	return !os.IsNotExist(err)
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

func ExpandDir(dir string) string {
	wd := filepath.Dir(os.Args[0])
	if filepath.IsAbs(dir) {
		return filepath.Clean(dir)
	}
	return filepath.Clean(filepath.Join(wd, dir))
}

func MkdirAll(path string) error {
	return os.MkdirAll(ExpandDir(path), 0o700)
}

func RemoveAll(path string) error {
	return os.RemoveAll(ExpandDir(path))
}
