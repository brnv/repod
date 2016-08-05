package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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

	permissionsPackageDefault = 0644
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
		if outputLine == "" {
			continue
		}

		packages = append(packages, outputLine)
	}

	return packages, nil
}

func (arch *RepositoryArch) ListEpoches() ([]string, error) {
	epochFiles, err := ioutil.ReadDir(arch.path)
	if err != nil {
		if os.IsNotExist(err) {
			err = errors.New("unknown repo")
		}
		return []string{}, err
	}

	epoches := []string{}
	for _, epochFile := range epochFiles {
		epoches = append(epoches, epochFile.Name())
	}

	return epoches, nil
}

func (arch *RepositoryArch) AddPackage(
	packageName string, packageFile io.Reader, force bool,
) error {
	packageFilePath := arch.getPackagesPath() + "/" + packageName

	contentRaw, err := ioutil.ReadAll(packageFile)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(
		packageFilePath,
		contentRaw,
		permissionsPackageDefault,
	)
	if err != nil {
		return err
	}

	cmd := exec.Command("gpg", "--detach-sign", "--yes", packageFilePath)
	_, _, err = executil.Run(cmd)
	if err != nil {
		return err
	}

	cmdOptions := []string{
		"-s",
		arch.getDatabaseFilePath(),
		packageFilePath,
	}
	if !force {
		cmdOptions = append([]string{"-n"}, cmdOptions...)
	}

	_, stderr, err := executil.Run(exec.Command("repo-add", cmdOptions...))
	if err != nil {
		return err
	}

	if !force && string(stderr) != "" {
		return fmt.Errorf("can't add package that already exists in repo")
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

func (arch *RepositoryArch) EditPackage(
	packageName string, file io.Reader,
) error {
	err := arch.AddPackage(packageName, file, true)
	if err != nil {
		return err
	}

	return nil
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

func (arch *RepositoryArch) GetPackageFile(
	packageName string,
) (io.Reader, error) {
	files, err := filepath.Glob(arch.getPackagesPath() + "/" + packageName)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf(
			"can't find package file for package '%s'",
			packageName,
		)
	}

	if len(files) > 1 {
		return nil, fmt.Errorf(
			"can't found one package file for package '%s', found: '%#v'",
			packageName, files,
		)
	}

	file, err := os.Open(files[0])
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (arch *RepositoryArch) SetEpoch(epoch string) {
	arch.epoch = epoch
}
