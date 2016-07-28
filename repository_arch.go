package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/kovetskiy/executil"
)

type RepositoryArch struct {
	Path         string
	Epoch        string
	Database     string
	Architecture string
}

const (
	formatPacmanConfRepo = "[%s]"
)

func (arch RepositoryArch) ListPackages() ([]string, error) {
	directory, err := arch.getTmpPacmanDBDir()
	if err != nil {
		return []string{}, err
	}

	config, err := arch.getTmpPacmanConfig()
	if err != nil {
		return []string{}, err
	}

	cmd := exec.Command(
		"pacman",
		"-Sl",
		"--config",
		config,
		"-b",
		directory,
	)
	pacmanPackagesRaw, _, err := executil.Run(cmd)
	if err != nil {
		return []string{}, err
	}

	var (
		pacmanPackages = strings.Split(string(pacmanPackagesRaw), "\n")
		packages       = []string{}
	)
	for _, pacmanPackage := range pacmanPackages {
		if strings.Count(pacmanPackage, " ") > 1 {
			packages = append(
				packages,
				strings.Split(pacmanPackage, " ")[1],
			)
		}
	}

	return packages, nil
}

func (arch *RepositoryArch) AddPackage(
	packageName string, packageFile io.Reader,
) error {
	// check if this version were installed
	// find and remove any other version of package
	// put new version into backup dir

	packageFilePath := arch.getPackagesPath() + "/" + packageName

	contentRaw, err := ioutil.ReadAll(packageFile)
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
		"repo-add", "-s", arch.getDatabaseFilePath(), packageFilePath,
	)
	_, _, err = executil.Run(cmd)
	if err != nil {
		return err
	}

	return nil
}

func (arch RepositoryArch) RemovePackage(packageName string) error {
	cmd := exec.Command(
		"repo-remove", arch.getDatabaseFilePath(), packageName,
	)
	_, _, err := executil.Run(cmd)
	if err != nil {
		return err
	}

	// TODO: remove file from filesystem

	return nil
}

func (arch RepositoryArch) EditPackage(packageName string) error {
	return nil
}

func (arch RepositoryArch) DescribePackage(packageName string) error {
	return nil
}

func (arch *RepositoryArch) getDatabaseName() string {
	return arch.Database + "-" + arch.Epoch
}

func (arch *RepositoryArch) getDatabaseFilename() string {
	return arch.getDatabaseName() + ".db"
}

func (arch *RepositoryArch) getDatabaseFilePath() string {
	return arch.getPackagesPath() + "/" +
		arch.getDatabaseFilename() + ".tar.xz"
}

func (arch RepositoryArch) getPackagesPath() string {
	return arch.Path + "/" + arch.Epoch + "/" +
		arch.Database + "/" + arch.Architecture
}

func (arch *RepositoryArch) getTmpPacmanDBDir() (string, error) {
	directory, err := ioutil.TempDir("/tmp/", "repod-")
	if err != nil {
		return "", err
	}

	err = os.Mkdir(directory+"/sync", 0777)
	if err != nil {
		return "", err
	}

	err = os.Symlink(
		arch.getDatabaseFilePath(),
		directory+"/sync/"+arch.getDatabaseFilename(),
	)
	if err != nil {
		return "", err
	}

	return directory, nil
}

func (arch *RepositoryArch) getTmpPacmanConfig() (string, error) {
	config, err := ioutil.TempFile("/tmp/", "repod-")
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(
		config.Name(),
		[]byte(fmt.Sprintf(formatPacmanConfRepo, arch.getDatabaseName())),
		0644,
	)

	if err != nil {
		return "", err
	}

	return config.Name(), nil
}
