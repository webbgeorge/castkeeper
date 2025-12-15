package config

import "os"

func MustOpenLocalFSRoot(basePath string) *os.Root {
	err := os.MkdirAll(basePath, 0750)
	if err != nil {
		panic(err)
	}
	root, err := os.OpenRoot(basePath)
	if err != nil {
		panic(err)
	}
	return root
}
