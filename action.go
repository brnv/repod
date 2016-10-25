package main

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/reconquest/ser-go"
)

func listRepositories(root string) ([]string, error) {
	debugf("reading root directory: '%s'", root)

	items, err := ioutil.ReadDir(root)
	if err != nil {
		return []string{}, ser.Errorf(
			err,
			"can't read root directory %s", root,
		)
	}

	repositories := []string{}
	for _, repository := range items {
		if !repository.IsDir() {
			continue
		}

		repositories = append(repositories, repository.Name())
	}

	tracef("repositories: %#v", repositories)

	return repositories, nil
}

func addPackage(
	repository Repository,
	packagePath string,
	force bool,
) error {
	debugf("opening package file '%s'", packagePath)

	file, err := os.Open(packagePath)
	if err != nil {
		return ser.Errorf(
			err,
			"can't open file %s", packagePath,
		)
	}

	debugf("copying file to repository directory")

	packageFilePath, err := repository.CreatePackageFile(
		path.Base(packagePath),
		file,
	)
	if err != nil {
		return ser.Errorf(
			err,
			"can't copy file %s to repository directory",
			path.Base(packagePath),
		)
	}

	tracef("copied file path: %s", packageFilePath)

	debugf("adding package to repository")

	err = repository.AddPackage(packageFilePath, force)
	if err != nil {
		return ser.Errorf(err, `can't add package %s`, packagePath)
	}

	return nil
}
