package main

import (
	"io/ioutil"
	"os"
	"path"

	ser "github.com/reconquest/ser-go"
	"github.com/seletskiy/hierr"
)

func listRepositories(root string) ([]string, error) {
	tracef("reading root directory")

	repositoriesFileInfo, err := ioutil.ReadDir(root)
	if err != nil {
		return []string{}, hierr.Errorf(
			err,
			`can't read directory %s`, root,
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
		return hierr.Errorf(err, `can't open given file`)
	}

	tracef("copying file to repository directory")

	pathCopied, err := repository.CopyFileToRepo(path.Base(packagePath), file)
	if err != nil {
		return hierr.Errorf(err, "can't copy file to destination")
	}

	debugf("copied file path: %s", pathCopied)

	tracef("adding package to repository")

	err = repository.AddPackage(pathCopied, forcePackageAdd)
	if err != nil {
		return hierr.Errorf(err, `can't add package`)
	}

	return nil
}

func describePackage(
	repository Repository, packageName string,
) (string, error) {
	tracef("getting information about package")

	description, err := repository.DescribePackage(packageName)
	if err != nil {
		return "", hierr.Errorf(err, `can't get package description`)
	}

	debugf("package description: %#v", description)

	return description, nil
}

func editPackage(
	repository Repository, packageName string,
	packagePath string, pathNew string,
) error {
	var (
		file *os.File
		err  error
	)

	if len(packagePath) > 0 {
		tracef("opening given package file")

		file, err = os.Open(packagePath)
		if err != nil {
			return ser.Errorf(err, "can't open file")
		}
	}

	if file == nil {
		tracef("getting package file from repository")

		file, err = repository.GetPackageFile(packageName)
		if err != nil {
			return ser.Errorf(
				err,
				"can't get package file from repository",
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
		return ser.Errorf(err, "can't add package")
	}

	return nil
}
