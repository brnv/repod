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
	"github.com/reconquest/ser-go"
)

type RepositoryArch struct {
	root string
	path string
}

const formatPacmanConfRepo = "[%s]"

func (arch RepositoryArch) ListPackages() ([]string, error) {
	directory, err := arch.getSyncDirectory()
	if err != nil {
		return []string{}, ser.Errorf(err, `can't get sync directory`)
	}

	debugf("sync directory: %s", directory)

	defer os.RemoveAll(directory)

	tracef("generating pacman config for sync")

	config, err := arch.getPacmanConfig()
	if err != nil {
		return []string{}, ser.Errorf(err, `can't get pacman config`)
	}

	debugf("config: %s", config)

	defer os.RemoveAll(config)

	tracef("executing pacman")

	args := []string{
		"--sync", "--list", "--debug",
		"--config", config, "--dbpath", directory,
	}

	stdout, stderr, err := executil.Run(
		exec.Command("pacman", args...),
	)
	if err != nil {
		return []string{}, ser.Errorf(err, `can't execute pacman command`)
	}

	if len(stdout) > 0 {
		debugf("pacman stdout: %s", stdout)
	}

	if len(stderr) > 0 {
		debugf("pacman stderr: %s", stderr)
	}

	packages := strings.Split(string(stdout), "\n")

	return packages, nil
}

func (arch *RepositoryArch) signUpPackage(packageName string) error {
	tracef("executing gpg")

	args := []string{
		"--detach-sign",
		"--yes",
		packageName,
	}

	stdout, stderr, err := executil.Run(
		exec.Command("gpg", args...),
	)
	if err != nil {
		return ser.Errorf(err, "can't execute gpg")
	}

	debugf("gpg stdout: %s", stdout)

	debugf("gpg stderr: %s", stderr)

	return nil
}

func (arch *RepositoryArch) addPackage(
	packageFile *os.File, force bool,
) error {
	tracef("executing repo-add")

	args := []string{
		"-s",
		arch.getDatabaseFilepath(),
		packageFile.Name(),
	}

	if !force {
		args = append([]string{"-n"}, args...)
	}

	stdout, stderr, err := executil.Run(
		exec.Command("repo-add", args...),
	)
	if err != nil {
		return ser.Errorf(err, "can't execute repo-add")
	}

	if len(stdout) > 0 {
		debugf("repo-add stdout: %s", stdout)
	}

	if len(stderr) > 0 {
		debugf("repo-add stderr: %s", stderr)

		if !force {
			return ser.Errorf(
				errors.New(string(stderr)),
				"repo-add errors",
			)
		}
	}

	return nil
}

func (arch *RepositoryArch) CreatePackageFile(
	packageName string, packageFile io.Reader,
) (string, error) {
	tracef("ensuring packages directory")

	packagesDir := arch.getPackagesPath()

	debugf("packages directory: %s", packagesDir)

	err := os.MkdirAll(packagesDir, 0644)
	if err != nil {
		return "", ser.Errorf(
			err,
			"can't create directory %s", packagesDir,
		)
	}

	tracef("creating package file")

	filepath := filepath.Join(packagesDir, packageName)

	debugf("package file: %s", filepath)

	file, err := os.Create(filepath)
	if err != nil {
		return "", ser.Errorf(err, "can't create file %s", filepath)
	}

	tracef("copying given file content to new file")

	_, err = io.Copy(file, packageFile)
	if err != nil {
		return "", ser.Errorf(
			err,
			"can't copy given file content to file %s", filepath,
		)
	}

	return file.Name(), nil
}

func (arch *RepositoryArch) AddPackage(
	packagePath string, force bool,
) error {
	tracef("reading given package file")

	file, err := os.Open(packagePath)
	if err != nil {
		return ser.Errorf(
			err,
			"can't read given package file %s", packagePath,
		)
	}

	tracef("signing up given package file")

	err = arch.signUpPackage(file.Name())
	if err != nil {
		return ser.Errorf(
			err,
			"can't sign up given package file %s", packagePath,
		)
	}

	tracef("updating repository")

	debugf("force mode: %#v", force)

	err = arch.addPackage(file, force)
	if err != nil {
		return ser.Errorf(
			err,
			"can't update repository with given package file %s",
			packagePath,
		)
	}

	return nil
}

func (arch *RepositoryArch) CopyPackage(
	packageName string,
	pathNew string,
) error {
	var (
		file *os.File
		err  error
	)

	tracef("getting package file from repository")

	file, err = arch.GetPackageFile(packageName)
	if err != nil {
		return ser.Errorf(
			err,
			"can't get package '%s' from repository", packageName,
		)
	}

	tracef("changing target repository path")

	arch.SetPath(pathNew)

	tracef("copying package file")

	err = arch.AddPackage(file.Name(), true)
	if err != nil {
		return ser.Errorf(
			err,
			"can't copy package %s to path %s", packageName, pathNew,
		)
	}

	return nil
}

func (arch RepositoryArch) RemovePackage(packageName string) error {
	tracef("executing repo-remove")

	args := []string{
		arch.getDatabaseFilepath(),
		packageName,
	}

	stdout, stderr, err := executil.Run(
		exec.Command("repo-remove", args...),
	)
	if err != nil {
		return ser.Errorf(err, "can't execute repo-remove")
	}

	debugf("repo-remove stdout: %s", stdout)

	debugf("repo-remove stderr: %s", stderr)

	tracef("getting package file from repository")

	file, err := arch.GetPackageFile(packageName)
	if err != nil {
		return ser.Errorf(
			err,
			"can't get file for package %s", packageName,
		)
	}

	tracef("removing file")

	err = os.Remove(file.Name())
	if err != nil {
		return ser.Errorf(
			err,
			"can't remove file %s", file.Name(),
		)
	}

	return nil
}

func (arch RepositoryArch) DescribePackage(
	packageName string,
) (string, error) {
	directory, err := arch.getSyncDirectory()
	if err != nil {
		return "", ser.Errorf(err, `can't get sync directory`)
	}

	defer os.RemoveAll(directory)

	config, err := arch.getPacmanConfig()
	if err != nil {
		return "", ser.Errorf(err, `can't get pacman config`)
	}

	debugf("pacman config: %s", config)

	defer os.RemoveAll(config)

	args := []string{
		"--sync", "--info",
		"--config", config,
		"--dbpath", directory, packageName,
	}

	tracef("executing pacman")

	stdout, stderr, err := executil.Run(
		exec.Command("pacman", args...),
	)
	if err != nil {
		return "", ser.Errorf(err, "can't execute pacman")
	}

	debugf("pacman stdout: %s", stdout)

	debugf("pacman stderr: %s", stderr)

	return string(stdout), nil
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

func (arch *RepositoryArch) getPacmanConfig() (string, error) {
	tracef("generating pacman config")

	tracef("making temporary file")

	config, err := ioutil.TempFile("/tmp/", "repod-pacman-config-")
	if err != nil {
		return "", ser.Errorf(err, "can't make temporary file")
	}

	debugf("temporary file: %s", config.Name())

	tracef("write config content to temporary file")

	content := fmt.Sprintf(formatPacmanConfRepo, arch.getDatabaseName())

	debugf("config file content: '%s'", content)

	err = ioutil.WriteFile(config.Name(), []byte(content), 0644)
	if err != nil {
		return "", ser.Errorf(
			err,
			"can't write file %s", config,
		)
	}

	return config.Name(), nil
}

func (arch *RepositoryArch) getSyncDirectory() (string, error) {
	tracef("getting sync directory")

	tracef("making temporary directory")

	directoryTemp, err := ioutil.TempDir("/tmp/", "repod-")
	if err != nil {
		return "", ser.Errorf(
			err,
			"can't make temporary directory",
		)
	}

	debugf("temporary directory: %s", directoryTemp)

	tracef("making sync directory")

	directorySync := directoryTemp + "/sync"

	err = os.Mkdir(directorySync, 0700)
	if err != nil {
		return "", ser.Errorf(
			err,
			"can't make sync directory %s", directorySync,
		)
	}

	debugf("sync directory: %s", directorySync)

	tracef("symlinking database file to sync directory")

	databaseFile := arch.getDatabaseFilepath()

	databaseFileSync := filepath.Join(
		directorySync,
		arch.getDatabaseFilename(),
	)

	debugf(
		"symlinking source: %s, destination: %s",
		databaseFile,
		databaseFileSync,
	)

	err = os.Symlink(databaseFile, databaseFileSync)
	if err != nil {
		return "", ser.Errorf(
			err,
			"can't symlink database file to directory %s",
			directorySync,
		)
	}

	return directoryTemp, nil
}

func (arch *RepositoryArch) GetPackageFile(
	packageName string,
) (*os.File, error) {
	tracef("finding package file in packages")

	glob := fmt.Sprintf(
		"%s-[0-9]*-[0-9]*-*.pkg.tar.xz",
		packageName,
	)

	debugf("pattern: %s", glob)

	files, err := filepath.Glob(
		filepath.Join(arch.getPackagesPath(), glob),
	)
	if err != nil {
		return nil, ser.Errorf(
			err,
			"can't find file in directory %s by pattern %s",
			arch.getPackagesPath(), glob,
		)
	}

	switch len(files) {
	case 0:
		return nil, ser.Errorf(
			err,
			"no files found by pattern %s",
			glob,
		)

	case 1:
		file, err := os.Open(files[0])
		if err != nil {
			return nil, ser.Errorf(
				err,
				"can't open package file %s", files[0],
			)
		}
		return file, nil

	default:
		return nil, ser.Errorf(err, "multiple files found by pattern")

	}
}

func (arch *RepositoryArch) SetPath(path string) {
	arch.path = path
}
