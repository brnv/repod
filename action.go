package main

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/kovetskiy/hierr"
)

func listRepositories(root string) (string, error) {
	repositoriesFileInfo, err := ioutil.ReadDir(root)
	if err != nil {
		return "", hierr.Errorf(
			err,
			`can't list repositories from root %s`, root,
		)
	}

	repositories := []string{}
	for _, repository := range repositoriesFileInfo {
		repositories = append(repositories, repository.Name())
	}

	return strings.Join(repositories, "\n"), nil
}

func listPackages(repository Repository) (string, error) {
	packages, err := repository.ListPackages()
	if err != nil {
		return "", hierr.Errorf(
			err,
			`can't list packages`,
		)
	}

	return strings.Join(packages, "\n"), nil
}

func addPackage(
	repository Repository,
	packagePath string,
) (string, error) {
	file, err := os.Open(packagePath)
	if err != nil {
		return "", hierr.Errorf(
			err,
			`can't open given file %s`, packagePath,
		)
	}

	_, err = repository.CopyFileToRepo(path.Base(file.Name()), file)
	if err != nil {
		return "", hierr.Errorf(
			err,
			"can't copy file to repository dir",
		)
	}

	err = repository.AddPackage(path.Base(packagePath), file, false)
	if err != nil {
		return "", hierr.Errorf(
			err,
			`can't add package from file %s`, packagePath,
		)
	}

	return "package was successfully added", nil
}

func describePackage(
	repository Repository, packageName string,
) (string, error) {
	description, err := repository.DescribePackage(packageName)
	if err != nil {
		return "", hierr.Errorf(
			err,
			`can't get package '%s' description`, packageName,
		)
	}

	return description, nil
}

func editPackage(
	repository Repository, packageName string,
	packageFile string, rootNew string,
) (string, error) {
	var (
		filename string
		file     *os.File
		err      error
	)

	if rootNew != "" {
		filename, file, err = repository.GetPackageFile(packageName)
		if err != nil {
			return "", hierr.Errorf(
				err,
				"can't get package file %s", packageName,
			)
		}
		repository.SetPath(rootNew)
	} else {
		file, err = os.Open(packageFile)
		if err != nil {
			return "", err
		}
		filename = file.Name()
	}

	err = repository.AddPackage(filename, file, true)
	if err != nil {
		return "", err
	}

	return "package was successfully edited", nil
}
