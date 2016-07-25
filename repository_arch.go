package main

import "io/ioutil"

type RepositoryArch struct {
	Path         string
	Epoch        string
	Database     string
	Architecture string
}

func (arch RepositoryArch) getPackagesPath() string {
	return arch.Path + "/" + arch.Epoch + "/" +
		arch.Database + "/" + arch.Architecture
}

func (arch RepositoryArch) ListPackages() ([]string, error) {
	packagesInfo, err := ioutil.ReadDir(arch.getPackagesPath())
	if err != nil {
		return []string{}, nil
	}

	packages := []string{}

	for _, packageInfo := range packagesInfo {
		packages = append(packages, packageInfo.Name())
	}

	return packages, nil
}

func (arch RepositoryArch) AddPackage(
	repositoryPackage RepositoryPackage,
) error {
	return nil
}

func (arch RepositoryArch) DeletePackage(
	repositoryPackage RepositoryPackage,
) error {
	return nil
}

func (arch RepositoryArch) EditPackage(
	repositoryPackage RepositoryPackage,
) error {
	return nil
}

func (arch RepositoryArch) DescribePackage(
	repositoryPackage RepositoryPackage,
) error {
	return nil
}
