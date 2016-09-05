package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type Repository interface {
	ListPackages() ([]string, error)
	ListEpoches() ([]string, error)

	AddPackage(packageName string, file *os.File, force bool) error
	CopyFileToRepo(packageName string, file io.Reader) (string, error)

	RemovePackage(packageName string) (string, error)
	DescribePackage(packageName string) (string, error)
	GetPackageFile(packageName string) (string, *os.File, error)

	SetPath(path string)
}

func getRepositorySystem(path string) string {
	switch {
	case strings.HasPrefix(path, "arch"):
		return "archlinux"

	case strings.HasPrefix(path, "debian"),
		strings.HasPrefix(path, "ubuntu"):
		return "debian"

	default:
		return ""
	}
}

func getRepository(root, path, system string) (Repository, error) {
	if system == "autodetect" {
		system = getRepositorySystem(path)
	}

	switch system {
	case "arch", "archlinux":
		return &RepositoryArch{
			root: root,
			path: path,
		}, nil

	case "ubuntu", "debian":
		panic("not implemented")

	default:
		return nil, fmt.Errorf(
			"can't obtain repository system, try to specify --system <type>",
			system,
		)
	}
}
