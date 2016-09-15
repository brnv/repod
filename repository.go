package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type Repository interface {
	ListPackages() ([]string, error)
	AddPackage(packagePath string, force bool) error
	CopyPackage(packageName string, pathNew string) error
	CreatePackageFile(packageName string, file io.Reader) (string, error)
	RemovePackage(packageName string) error
	DescribePackage(packageName string) (string, error)
	GetPackageFile(packageName string) (*os.File, error)
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
		tracef("trying to detect repository type")
		system = getRepositorySystem(path)
		debugf("repository type: %s", system)
	}

	switch system {
	case "arch", "archlinux":
		return &RepositoryArch{root: root, path: path}, nil

	case "ubuntu", "debian":
		return nil, fmt.Errorf("system '%s' not implemented", system)

	default:
		return nil, fmt.Errorf(
			"can't obtain repository system, try to specify --system <type>",
			system,
		)
	}
}
