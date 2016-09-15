package main

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/seletskiy/hierr"
)

func listRepositories(root string) ([]string, error) {
	tracef("reading root directory")

	repositoriesFileInfo, err := ioutil.ReadDir(root)
	if err != nil {
		return []string{}, hierr.Errorf(
			err,
			"can't read root directory %s", root,
		)
	}

	repositories := []string{}
	for _, repository := range repositoriesFileInfo {
		repositories = append(repositories, repository.Name())
	}

	debugf("repositories: %#v", repositories)

	return repositories, nil
}

func addPackage(
	repository Repository,
	packagePath string,
	force bool,
) error {
	tracef("opening given package file")

	file, err := os.Open(packagePath)
	if err != nil {
		return hierr.Errorf(
			err,
			"can't open given file %s", packagePath,
		)
	}

	tracef("copying file to repository directory")

	packageFilePath, err := repository.CreatePackageFile(
		path.Base(packagePath),
		file,
	)
	if err != nil {
		return hierr.Errorf(
			err,
			"can't copy file %s to repository directory",
			path.Base(packagePath),
		)
	}

	debugf("copied file path: %s", packageFilePath)

	tracef("adding package to repository")

	err = repository.AddPackage(packageFilePath, force)
	if err != nil {
		return hierr.Errorf(err, `can't add package %s`, packagePath)
	}

	return nil
}
