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
	tracef("getting sync directory to list packages")

	directory, err := arch.getSyncDirectory()
	if err != nil {
		return []string{}, ser.Errorf(err, `can't get sync directory`)
	}

	defer os.RemoveAll(directory)

	tracef("generating pacman config for sync")

	config, err := arch.getPacmanConfig()
	if err != nil {
		return []string{}, ser.Errorf(err, `can't get pacman config`)
	}

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
		return ser.Errorf(err, "error while executing gpg")
	}

	debugf("gpg stdout: %s", stdout)

	debugf("gpg stderr: %s", stderr)

	return nil
}

func (arch *RepositoryArch) updateRepo(
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
		return ser.Errorf(err, "can't add package to repo")
	}

	if len(stdout) > 0 {
		debugf("repo-add stdout: %s", stdout)
	}

	if len(stderr) > 0 {
		debugf("repo-add stderr: %s", stderr)

		if !force {
			return ser.Errorf(
				errors.New(string(stderr)),
				"repo-add exec error",
			)
		}
	}

	return nil
}

func (arch *RepositoryArch) CreatePackageFile(
	packageName string, packageFile io.Reader,
) (string, error) {
	packagesDir := arch.getPackagesPath()

	tracef("ensuring packages directory")

	err := os.MkdirAll(packagesDir, 0644)
	if err != nil {
		return "", ser.Errorf(
			err, "can't create directory",
		)
	}

	tracef("creating package file")

	filepath := filepath.Join(packagesDir, packageName)

	file, err := os.Create(filepath)
	if err != nil {
		return "", ser.Errorf(err, "can't create file %s", filepath)
	}

	tracef("copying given file content to new file")

	_, err = io.Copy(file, packageFile)
	if err != nil {
		return "", ser.Errorf(err, "can't copy content to new file")
	}

	return file.Name(), nil
}

func (arch *RepositoryArch) AddPackage(
	packagePath string, force bool,
) error {
	tracef("opening given package file")

	file, err := os.Open(packagePath)
	if err != nil {
		return ser.Errorf(err, "can't open given package file")
	}

	tracef("signing up package")

	err = arch.signUpPackage(file.Name())
	if err != nil {
		return ser.Errorf(err, "can't sign up given package")
	}

	tracef("updating repository")

	debugf("force mode: %#v", force)

	err = arch.updateRepo(file, force)
	if err != nil {
		return ser.Errorf(err, "can't update repository")
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
		return ser.Errorf(err, "error while executing repo remove")
	}

	debugf("repo-remove stdout: %s", stdout)

	debugf("repo-remove stderr: %s", stderr)

	tracef("getting package file path")

	file, err := arch.GetPackageFile(packageName)
	if err != nil {
		return ser.Errorf(err, "can't get package file")
	}

	tracef("removing file")

	err = os.Remove(file.Name())
	if err != nil {
		return ser.Errorf(err, "can't remove file")
	}

	return nil
}

func (arch RepositoryArch) DescribePackage(
	packageName string,
) (string, error) {
	tracef("getting sync directory to describe package")

	directory, err := arch.getSyncDirectory()
	if err != nil {
		return "", err
	}

	debugf("sync directory: %s", directory)

	defer os.RemoveAll(directory)

	tracef("generating pacman config for sync")

	config, err := arch.getPacmanConfig()
	if err != nil {
		return "", ser.Errorf(
			err,
			`can't get pacman temporary config`,
		)
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
		return "", ser.Errorf(
			err,
			"pacman execution failed",
		)
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
	tracef("making temporary directory")

	directoryTemp, err := ioutil.TempDir("/tmp/", "repod-")
	if err != nil {
		return "", ser.Errorf(err, "can't get temporary directory")
	}

	debugf("temporary directory: %s", directoryTemp)

	tracef("making sync directory")

	directorySync := directoryTemp + "/sync"

	err = os.Mkdir(directorySync, 0777)
	if err != nil {
		return "", ser.Errorf(err, "can't make sync directory")
	}

	debugf("sync directory: %s", directorySync)

	tracef("symlinking database file to sync directory")

	err = os.Symlink(
		arch.getDatabaseFilepath(),
		filepath.Join(directorySync, arch.getDatabaseFilename()),
	)
	if err != nil {
		return "", ser.Errorf(err, "can't symlink database file")
	}

	return directoryTemp, nil
}

func (arch *RepositoryArch) GetPackageFile(
	packageName string,
) (*os.File, error) {
	tracef("searching files")

	glob := fmt.Sprintf(
		"%s/%s-[0-9]*-[0-9]*-*.pkg.tar.xz",
		arch.getPackagesPath(),
		packageName,
	)

	debugf("pattern: %s", glob)

	files, err := filepath.Glob(glob)
	if err != nil {
		return nil, ser.Errorf(err, `error while files searching`)
	}

	switch len(files) {
	case 0:
		return nil, ser.Errorf(err, "no files found by pattern")

	case 1:
		file, err := os.Open(files[0])
		if err != nil {
			return nil, ser.Errorf(err, "can't open given package file")
		}
		return file, nil

	default:
		return nil, ser.Errorf(err, "multiple files found by pattern")

	}
}

func (arch *RepositoryArch) SetPath(path string) {
	arch.path = path
}
