package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
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
	repo string, repoRoot string, epoch string, database string, architecture string,
) (Repository, error) {
	var (
		osType   = detectRepositoryOS(repo)
		repoPath = filepath.Join(repoRoot, repo)
	)

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
