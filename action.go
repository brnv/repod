package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/kovetskiy/hierr"
)

func listRepositories(repoRoot string) ([]string, error) {
	repositoriesFileInfo, err := ioutil.ReadDir(repoRoot)
	if err != nil {
		return []string{}, hierr.Errorf(
			err,
			`can't list repositories from root %s`, repoRoot,
		)
	}

	repositories := []string{}
	for _, repository := range repositoriesFileInfo {
		repositories = append(repositories, repository.Name())
	}

	return repositories, nil
}

func listEpoches(
	repoRoot string,
	repository Repository,
) (string, error) {
	epoches, err := repository.ListEpoches()
	if err != nil {
		return "", hierr.Errorf(
			err,
			`can't list epoches for repository`,
		)
	}

	if len(epoches) == 0 {
		return "", fmt.Errorf("no epoches found", nil)
	}

	return strings.Join(epoches, "\n"), nil
}

func listPackages(
	repoRoot string,
	repository Repository,
) (string, error) {
	packages, err := repository.ListPackages()
	if err != nil {
		return "", hierr.Errorf(
			err,
			`can't list packages`,
		)
	}

	if len(packages) == 0 {
		return "", fmt.Errorf("no packages found", nil)
	}

	return strings.Join(packages, "\n"), nil
}

func addPackage(
	repoRoot string, repository Repository,
	packageName string, packageFile string,
) (string, error) {
	file, err := os.Open(packageFile)
	if err != nil {
		return "", hierr.Errorf(
			err,
			`can't open given file %s`, packageFile,
		)
	}

	err = repository.AddPackage(packageName, file, false)
	if err != nil {
		return "", hierr.Errorf(
			err,
			`can't add given package`,
		)
	}

	return "package was successfully added", nil
}

func describePackage(
	repoRoot string, repository Repository, packageName string,
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
	repoRoot string, repository Repository,
	packageName string, packageFile string,
	epochToChange string,
) (string, error) {
	var (
		file io.Reader
		err  error
	)

	if epochToChange != "" {
		file, err = repository.GetPackageFile(packageName)
		if err != nil {
			return "", err
		}

		repository.SetEpoch(epochToChange)
	} else {
		file, err = os.Open(packageFile)
		if err != nil {
			return "", err
		}
	}

	err = repository.EditPackage(packageName, file)
	if err != nil {
		return "", err
	}

	return "package was successfully edited", nil
}
