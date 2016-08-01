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
	path         string
	epoch        string
	database     string
	architecture string
}

const (
	formatPacmanConfRepo = "[%s]"
)

func (arch RepositoryArch) ListPackages() ([]string, error) {
	directory, config, err := arch.preparePacmanDB()
	if err != nil {
		return []string{}, err
	}
	defer func() {
		os.RemoveAll(directory)
		os.RemoveAll(config)
	}()

	cmd := exec.Command(
		"pacman", "--sync", "--list",
		"--config", config,
		"--dbpath", directory,
	)
	pacmanOutput, _, err := executil.Run(cmd)
	if err != nil {
		return []string{}, err
	}

	var (
		outputLines = strings.Split(string(pacmanOutput), "\n")
		packages    = []string{}
	)
	for _, outputLine := range outputLines {
		// outputLine example: "testing-db-testing package_name 1-1"
		if strings.Count(outputLine, " ") > 1 {
			packages = append(
				packages,
				strings.Split(outputLine, " ")[1],
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

func (arch RepositoryArch) DescribePackage(
	packageName string,
) ([]string, error) {
	directory, config, err := arch.preparePacmanDB()
	if err != nil {
		return []string{}, err
	}
	defer func() {
		os.RemoveAll(directory)
		os.RemoveAll(config)
	}()

	cmd := exec.Command(
		"pacman", "--sync", "--info",
		"--config", config,
		"--dbpath", directory,
		packageName,
	)
	pacmanOutput, _, err := executil.Run(cmd)
	if err != nil {
		return []string{}, err
	}

	var (
		outputLines = strings.Split(string(pacmanOutput), "\n")
		packageInfo = []string{}
	)
	for _, outputLine := range outputLines {
		packageInfo = append(
			packageInfo,
			outputLine,
		)
	}

	return packageInfo, nil
}

func (arch *RepositoryArch) getDatabaseName() string {
	return arch.database + "-" + arch.epoch
}

func (arch *RepositoryArch) getDatabaseFilename() string {
	return arch.getDatabaseName() + ".db"
}

func (arch *RepositoryArch) getDatabaseFilePath() string {
	return arch.getPackagesPath() + "/" +
		arch.getDatabaseFilename() + ".tar.xz"
}

func (arch RepositoryArch) getPackagesPath() string {
	return arch.path + "/" + arch.epoch + "/" +
		arch.database + "/" + arch.architecture
}

func (arch *RepositoryArch) getTmpDirectory() (string, error) {
	directory, err := ioutil.TempDir("/tmp/", "repod-")
	if err != nil {
		return "", err
	}

	return directory, nil
}

func (arch *RepositoryArch) prepareSyncDirectory(directory string) error {
	syncDirectoryPath := directory + "/sync"

	err := os.Mkdir(syncDirectoryPath, 0777)
	if err != nil {
		return err
	}

	err = os.Symlink(
		arch.getDatabaseFilePath(),
		syncDirectoryPath+"/"+arch.getDatabaseFilename(),
	)
	if err != nil {
		return err
	}

	return nil
}

func (arch *RepositoryArch) getTmpPacmanConfig() (string, error) {
	config, err := ioutil.TempFile("/tmp/", "repod-pacman-config-")
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

func (arch *RepositoryArch) preparePacmanDB() (string, string, error) {
	directory, err := arch.getTmpDirectory()
	if err != nil {
		return "", "", err
	}

	err = arch.prepareSyncDirectory(directory)
	if err != nil {
		os.RemoveAll(directory)
		return "", "", err
	}

	config, err := arch.getTmpPacmanConfig()
	if err != nil {
		os.RemoveAll(directory)
		return "", "", err
	}

	return directory, config, nil
}
