package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

type Repository interface {
	ListPackages() ([]string, error)
	ListEpoches() ([]string, error)
	AddPackage(packageName string, file io.Reader, force bool) error
	RemovePackage(packageName string) error
	DescribePackage(packageName string) (string, error)
	EditPackage(packageName string, file io.Reader) error
	GetPackageFile(packageName string) (io.Reader, error)
	SetEpoch(epoch string)
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

func getRepository(
	osType string, repoPath string, epoch string, database string, architecture string,
) (Repository, error) {
	switch osType {
	case osArchLinux:
		return &RepositoryArch{
			path:         repoPath,
			epoch:        epoch,
			database:     database,
			architecture: architecture,
		}, nil

	default:
		return nil, fmt.Errorf(
			"repo manager for %s is not implemented",
			osType,
		)
	}
}

func listRepositories(root string) ([]string, error) {
	repositoriesFileInfo, err := ioutil.ReadDir(root)
	if err != nil {
		return []string{}, err
	}

	repositories := []string{}
	for _, repository := range repositoriesFileInfo {
		repositories = append(repositories, repository.Name())
	}

	return repositories, nil
}
