package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kovetskiy/hierr"
)

func getListRepositoriesOutput(repoRoot string) (string, error) {
	repositories, err := listRepositories(repoRoot)
	if err != nil {
		return "", hierr.Errorf(
			err,
			`can't list repositories from root %s`, repoRoot,
		)
	}

	if len(repositories) == 0 {
		return "", fmt.Errorf("no repos found", nil)
	}

	return strings.Join(repositories, "\n"), nil
}

func getListEpochesOutput(
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

func getListPackagesOutput(
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

func getAddPackageOutput(
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

func getDescribePackageOutput(
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

func getEditPackageOutput(
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
