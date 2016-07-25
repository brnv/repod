package main

import (
	"io/ioutil"
	"os/exec"

	"github.com/kovetskiy/executil"
)

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

func (arch *RepositoryArch) getDatabaseFilePath() string {
	return arch.getPackagesPath() + "/" +
		arch.Database + "-" + arch.Epoch + ".db.tar.xz"
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

func (arch *RepositoryArch) AddPackage(
	repositoryPackage RepositoryPackage,
) error {
	// check if this version were installed
	// find and remove any other version of package
	// put new version into backup dir

	packageFilePath := arch.getPackagesPath() + "/" + repositoryPackage.Name

	contentRaw, err := ioutil.ReadAll(repositoryPackage.File)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(packageFilePath, contentRaw, 0644)
	if err != nil {
		return err
	}

	cmd := exec.Command("gpg", "--detach-sign", "--yes", packageFilePath)
	_, _, err = executil.Run(cmd)
	if err != nil {
		return err
	}

	cmd = exec.Command(
		"repo-add",
		"-s",
		arch.getDatabaseFilePath(),
		packageFilePath,
	)
	_, _, err = executil.Run(cmd)
	if err != nil {
		return err
	}

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
