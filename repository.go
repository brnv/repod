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
	CopyPackage(name string, version string, pathNew string) error
	CreatePackageFile(name string, file io.Reader) (string, error)
	RemovePackage(name string, version string) error
	DescribePackage(name string) (string, error)
	GetPackageFile(name string, version string) (*os.File, error)
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
		debugf("trying to detect repository type")
		system = getRepositorySystem(path)
		tracef("repository type: '%s'", system)
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
