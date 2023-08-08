package utils

import (
	"os"
	"path/filepath"
)

func ExeDir() string {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(exe)
}
