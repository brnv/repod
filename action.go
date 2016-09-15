package main

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/reconquest/ser-go"
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

func addPackage(repository Repository, packagePath string) error {
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

	err = repository.AddPackage(packageFilePath, forcePackageAdd)
	if err != nil {
		return hierr.Errorf(err, `can't add package %s`, packagePath)
	}

	return nil
}

func editPackage(
	repository Repository, packageName string,
	packagePath string, pathNew string,
) error {
	var (
		file *os.File
		err  error
	)

	if len(packagePath) > 0 && file == nil {
		tracef("trying to open package file")

		file, err = os.Open(packagePath)
		if err != nil {
			return ser.Errorf(
				err,
				"can't open file %s", packagePath,
			)
		}
	}

	if file == nil {
		tracef("getting package file from repository")

		file, err = repository.GetPackageFile(packageName)
		if err != nil {
			return ser.Errorf(
				err,
				"can't get package '%s' from repository", packageName,
			)
		}
	}

	if len(pathNew) > 0 {
		tracef("changing target repository path")

		repository.SetPath(pathNew)
	}

	tracef("editing package file")

	err = repository.AddPackage(file.Name(), forcePackageEdit)
	if err != nil {
		return ser.Errorf(
			err,
			"can't edit package %s", packagePath,
		)
	}

	return nil
}
