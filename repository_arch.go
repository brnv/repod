package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kovetskiy/executil"
	"github.com/reconquest/hierr"
)

type RepositoryArch struct {
	path         string
	epoch        string
	database     string
	architecture string
}

const formatPacmanConfRepo = "[%s]"

func (arch RepositoryArch) ListPackages() ([]string, error) {
	directory, config, err := arch.preparePacmanSyncDir()
	if err != nil {
		return []string{}, hierr.Errorf(
			err,
			`can't prepare pacman sync directory`,
		)
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
		return []string{}, hierr.Errorf(
			err,
			`can't execute command %s`, cmd,
		)
	}

	packages := strings.Split(string(pacmanOutput), "\n")

	return packages[:len(packages)-1], nil
}

func (arch *RepositoryArch) ListEpoches() ([]string, error) {
	epochFiles, err := ioutil.ReadDir(arch.path)
	if err != nil {
		return []string{}, hierr.Errorf(
			err,
			`can't read dir %s`, arch.path,
		)
	}

	epoches := []string{}
	for _, epochFile := range epochFiles {
		epoches = append(epoches, epochFile.Name())
	}

	return epoches, nil
}

func (arch *RepositoryArch) signUpPackage(packageFilename string) error {
	cmd := exec.Command(
		"gpg", "--detach-sign", "--yes", packageFilename,
	)
	_, _, err := executil.Run(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (arch *RepositoryArch) putPackageIntoRepo(
	packageFilename string, packageFile io.Reader, force bool,
) error {
	cmdOptions := []string{
		"-s",
		arch.getDatabaseFilepath(),
		packageFilename,
	}
	if !force {
		cmdOptions = append([]string{"-n"}, cmdOptions...)
	}

	_, stderr, err := executil.Run(
		exec.Command("repo-add", cmdOptions...),
	)
	if err != nil {
		return hierr.Errorf(
			err,
			"can't add package to repo, exec args: %#v", cmdOptions,
		)
	}

	if !force && string(stderr) != "" {
		return fmt.Errorf(
			"repo-add exec error, args: %#v, stderr: %s",
			cmdOptions,
			string(stderr),
		)
	}

	return nil
}

func (arch *RepositoryArch) PutFileToRepo(
	packageFilename string, packageFile io.Reader,
) error {
	dstPackageFile, err := os.Create(
		filepath.Join(arch.getPackagesPath(), packageFilename),
	)
	if err != nil {
		return err
	}

	_, err = io.Copy(dstPackageFile, packageFile)
	if err != nil {
		return err
	}

	return nil
}

func (arch *RepositoryArch) AddPackage(
	packageFilename string, packageFile io.Reader, force bool,
) error {
	err := arch.signUpPackage(packageFilename)
	if err != nil {
		return err
	}

	return arch.putPackageIntoRepo(packageFilename, packageFile, force)
}

func (arch RepositoryArch) RemovePackage(packageName string) error {
	cmd := exec.Command(
		"repo-remove", arch.getDatabaseFilepath(), packageName,
	)
	_, _, err := executil.Run(cmd)
	if err != nil {
		return err
	}

	pattern := filepath.Join(arch.getPackagesPath(), packageName) + "*.tar.xz"

	files, err := filepath.Glob(pattern)
	if err != nil {
		return hierr.Errorf(
			err,
			`can't find files by pattern %s`, pattern,
		)
	}

	if len(files) == 0 {
		return fmt.Errorf(
			"no packages find to remove",
		)
	}

	for _, file := range files {
		err = os.Remove(file)
		if err != nil {
			return hierr.Errorf(
				err,
				"can't delete package file from filesystem %s",
				file,
			)
		}
	}

	return nil
}

func (arch RepositoryArch) DescribePackage(
	packageName string,
) (string, error) {
	directory, config, err := arch.preparePacmanSyncDir()
	if err != nil {
		return "", err
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
		return "", err
	}

	return string(pacmanOutput), nil
}

func (arch *RepositoryArch) getDatabaseName() string {
	return arch.database + "-" + arch.epoch
}

func (arch *RepositoryArch) getDatabaseFilename() string {
	return arch.getDatabaseName() + ".db"
}

func (arch *RepositoryArch) getDatabaseFilepath() string {
	return filepath.Join(
		arch.getPackagesPath(), arch.getDatabaseFilename()+".tar.xz",
	)
}

func (arch RepositoryArch) getPackagesPath() string {
	return filepath.Join(
		arch.path,
		arch.epoch,
		arch.database,
		arch.architecture,
	)
}

func (arch *RepositoryArch) prepareSyncDirectory(directory string) error {
	syncDirectoryPath := directory + "/sync"

	err := os.Mkdir(syncDirectoryPath, 0777)
	if err != nil {
		return hierr.Errorf(err, "can't create dir %s", syncDirectoryPath)
	}

	targetLink := filepath.Join(syncDirectoryPath, arch.getDatabaseFilename())
	err = os.Symlink(
		arch.getDatabaseFilepath(),
		targetLink,
	)
	if err != nil {
		return hierr.Errorf(
			err,
			"can't symlink %s to %s", arch.getDatabaseFilepath(), targetLink,
		)
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

func (arch *RepositoryArch) preparePacmanSyncDir() (string, string, error) {
	directory, err := ioutil.TempDir("/tmp/", "repod-")
	if err != nil {
		return "", "", hierr.Errorf(err, "can't create temp dir: %s")
	}

	err = arch.prepareSyncDirectory(directory)
	if err != nil {
		os.RemoveAll(directory)
		return "", "", hierr.Errorf(err, `can't prepare pacman sync directory`)
	}

	config, err := arch.getTmpPacmanConfig()
	if err != nil {
		os.RemoveAll(directory)
		return "", "", hierr.Errorf(err, `can't get pacman temporary config`)
	}

	return directory, config, nil
}

func (arch *RepositoryArch) GetPackageFile(
	packageName string,
) (string, *os.File, error) {
	pattern := packageName + "*.tar.xz"

	files, err := filepath.Glob(pattern)
	if err != nil {
		return "", nil, hierr.Errorf(
			err,
			`can't find files by pattern %s`, pattern,
		)
	}

	if len(files) == 0 {
		return "", nil, fmt.Errorf("no files found by pattern %s", pattern)
	}

	if len(files) > 1 {
		return "", nil, fmt.Errorf(
			"more than one file found by pattern %s, %#v",
			pattern, files,
		)
	}

	file, err := os.Open(files[0])
	if err != nil {
		return "", nil, hierr.Errorf(err, `can't open file %s`, files[0])
	}

	return file.Name(), file, nil
}

func (arch *RepositoryArch) SetEpoch(epoch string) {
	arch.epoch = epoch
}
