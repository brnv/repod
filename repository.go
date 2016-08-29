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

	RemovePackage(packageName string) error
	DescribePackage(packageName string) (string, error)
	GetPackageFile(packageName string) (string, *os.File, error)

	SetPath(path string)
}

const (
	osArchLinux = "arch"
	osUbuntu    = "ubuntu"
	osUnknown   = "unknown"
)

func detectRepositoryOS(repository string) string {
	if strings.HasPrefix(repository, "arch") {
		return osArchLinux
	}

	if strings.HasPrefix(repository, "ubuntu") {
		return osUbuntu
	}

	return osUnknown
}

func getRepository(root string, path string) (Repository, error) {
	osType := detectRepositoryOS(path)

	switch osType {
	case osArchLinux:
		return &RepositoryArch{
			root: root,
			path: path,
		}, nil

	default:
		return nil, fmt.Errorf(
			"repo manager for %s is not implemented",
			osType,
		)
	}
}
