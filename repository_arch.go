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
	root string
	path string
}

const formatPacmanConfRepo = "[%s]"

func (arch RepositoryArch) ListPackages() ([]string, error) {
	directory, err := arch.getSyncDirectory()
	if err != nil {
		return []string{}, hierr.Errorf(
			err,
			`can't prepare pacman sync directory`,
		)
	}

	config, err := arch.getPacmanConfig()
	if err != nil {
		os.RemoveAll(directory)
		return []string{}, hierr.Errorf(
			err,
			`can't get pacman temporary config`,
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

func (arch *RepositoryArch) signUpPackage(
	packageName string,
) error {
	cmd := exec.Command(
		"gpg", "--detach-sign", "--yes",
		packageName,
	)

	_, _, err := executil.Run(cmd)
	if err != nil {
		return err
	}

	return nil
}

func (arch *RepositoryArch) updateRepo(
	packageName string, packageFile io.Reader, force bool,
) error {
	cmdOptions := []string{
		"-s",
		arch.getDatabaseFilepath(),
		packageName,
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

func (arch *RepositoryArch) CopyFileToRepo(
	packageName string, packageFile io.Reader,
) (string, error) {
	dstPackagePath := filepath.Join(arch.getPackagesPath(), packageName)

	err := os.MkdirAll(filepath.Dir(dstPackagePath), 0644)
	if err != nil {
		return "", hierr.Errorf(
			err, "can't create directory %s", filepath.Dir(dstPackagePath),
		)
	}

	dstPackageFile, err := os.Create(dstPackagePath)
	if err != nil {
		return "", hierr.Errorf(
			err,
			"can't create package file %s", packageName,
		)
	}

	_, err = io.Copy(dstPackageFile, packageFile)
	if err != nil {
		return "", hierr.Errorf(
			err,
			"can't copy file content to destination",
		)
	}

	return dstPackageFile.Name(), nil
}

func (arch *RepositoryArch) AddPackage(
	packageName string, packageFile *os.File, force bool,
) error {
	err := arch.signUpPackage(packageFile.Name())
	if err != nil {
		return err
	}

	err = arch.updateRepo(packageName, packageFile, force)
	if err != nil {
		return err
	}

	return nil
}

func (arch RepositoryArch) RemovePackage(packageName string) error {
	cmd := exec.Command(
		"repo-remove", arch.getDatabaseFilepath(), packageName,
	)
	_, _, err := executil.Run(cmd)
	if err != nil {
		return err
	}

	searchPattern := filepath.Join(
		arch.getPackagesPath(), packageName,
	) + "*.tar.xz"

	files, err := filepath.Glob(searchPattern)
	if err != nil {
		return hierr.Errorf(
			err,
			`can't find files by searchPattern %s`, searchPattern,
		)
	}

	if len(files) == 0 {
		return fmt.Errorf("no packages found to remove")
	}

	for _, file := range files {
		err = os.Remove(file)
		if err != nil {
			return hierr.Errorf(
				err,
				"can't remove package file %s",
				file,
			)
		}
	}

	return nil
}

func (arch RepositoryArch) DescribePackage(
	packageName string,
) (string, error) {
	directory, err := arch.getSyncDirectory()
	if err != nil {
		return "", err
	}

	config, err := arch.getPacmanConfig()
	if err != nil {
		os.RemoveAll(directory)
		return "", hierr.Errorf(
			err,
			`can't get pacman temporary config`,
		)
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
	return strings.Replace(arch.path, "/", "-", -1)
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
	return filepath.Join(arch.root, arch.path)
}

func (arch *RepositoryArch) prepareSyncDirectory(directory string) error {
	syncDirectoryPath := directory + "/sync"

	err := os.Mkdir(syncDirectoryPath, 0777)
	if err != nil {
		return hierr.Errorf(
			err,
			"can't create dir %s", syncDirectoryPath,
		)
	}

	databaseLinkPath := filepath.Join(
		syncDirectoryPath,
		arch.getDatabaseFilename(),
	)

	err = os.Symlink(
		arch.getDatabaseFilepath(),
		databaseLinkPath,
	)
	if err != nil {
		return hierr.Errorf(
			err,
			"can't symlink %s to %s",
			arch.getDatabaseFilepath(), databaseLinkPath,
		)
	}

	return nil
}

func (arch *RepositoryArch) getPacmanConfig() (string, error) {
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

func (arch *RepositoryArch) getSyncDirectory() (string, error) {
	directory, err := ioutil.TempDir("/tmp/", "repod-")
	if err != nil {
		return "", hierr.Errorf(
			err,
			"can't create temp dir: %s",
		)
	}

	err = arch.prepareSyncDirectory(directory)
	if err != nil {
		os.RemoveAll(directory)
		return "", hierr.Errorf(
			err,
			`can't prepare pacman sync directory`,
		)
	}

	return directory, nil
}

func (arch *RepositoryArch) GetPackageFile(
	packageName string,
) (string, *os.File, error) {
	searchPattern := packageName + "*.tar.xz"

	files, err := filepath.Glob(searchPattern)
	if err != nil {
		return "", nil, hierr.Errorf(
			err,
			`can't find files by searchPattern %s`, searchPattern,
		)
	}

	if len(files) == 0 {
		return "", nil, fmt.Errorf(
			"no files found by searchPattern %s",
			searchPattern,
		)
	}

	if len(files) > 1 {
		return "", nil, fmt.Errorf(
			"more than one file found by searchPattern %s, %#v",
			searchPattern, files,
		)
	}

	file, err := os.Open(files[0])
	if err != nil {
		return "", nil, hierr.Errorf(err, `can't open file %s`, files[0])
	}

	return file.Name(), file, nil
}

func (arch *RepositoryArch) SetPath(path string) {
	arch.path = path
}
