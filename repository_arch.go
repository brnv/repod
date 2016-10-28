package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
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

	tracef("sync directory: %s", directory)

	defer func() {
		err = os.RemoveAll(directory)
		if err != nil {
			warningf("can't remove sync directory: %s", err)
		}
	}()

	debugf("generating pacman config for sync")

	config, err := arch.getPacmanConfig()
	if err != nil {
		return []string{}, ser.Errorf(err, `can't get pacman config`)
	}

	tracef("config: %s", config)

	defer func() {
		err = os.RemoveAll(config)
		if err != nil {
			warningf("can't remove pacman config: %s", err)
		}
	}()

	debugf("executing pacman")

	stdout, stderr, err := executil.Run(
		exec.Command("pacman", []string{
			"--sync", "--list", "--debug",
			"--config", config, "--dbpath", directory,
		}...),
	)
	if err != nil {
		return []string{}, ser.Errorf(err, `can't execute pacman command`)
	}

	tracef("pacman stdout: %s", stdout)
	tracef("pacman stderr: %s", stderr)

	packages := strings.Split(string(stdout), "\n")

	return packages, nil
}

func (arch *RepositoryArch) signUpPackage(packageName string) error {
	debugf("executing gpg")

	stdout, stderr, err := executil.Run(
		exec.Command("gpg", []string{
			"--detach-sign",
			"--yes",
			packageName,
		}...),
	)
	if err != nil {
		return ser.Errorf(err, "can't execute gpg")
	}

	tracef("gpg stdout: %s", stdout)
	tracef("gpg stderr: %s", stderr)

	return nil
}

func (arch *RepositoryArch) addPackage(
	packageFile *os.File, force bool,
) error {
	debugf("executing repo-add")

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

	tracef("repo-add stdout: %s", stdout)

	if len(stderr) > 0 {
		tracef("repo-add stderr: %s", stderr)

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
	debugf("ensuring packages directory")

	packagesDir := arch.getPackagesPath()

	tracef("packages directory: %s", packagesDir)

	err := os.MkdirAll(packagesDir, 0644)
	if err != nil {
		return "", ser.Errorf(
			err,
			"can't create directory %s", packagesDir,
		)
	}

	debugf("creating package file")

	filepath := filepath.Join(packagesDir, packageName)

	tracef("package file: %s", filepath)

	file, err := os.Create(filepath)
	if err != nil {
		return "", ser.Errorf(err, "can't create file %s", filepath)
	}

	debugf("copying given file content to new file")

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
	debugf("opening given package file")

	file, err := os.Open(packagePath)
	if err != nil {
		return ser.Errorf(
			err,
			"can't open given package file %s", packagePath,
		)
	}

	debugf("signing up given package file")

	err = arch.signUpPackage(file.Name())
	if err != nil {
		return ser.Errorf(
			err,
			"can't sign up given package file %s", packagePath,
		)
	}

	debugf("updating repository")

	tracef("force mode: %#v", force)

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
	name string,
	version string,
	pathNew string,
	force bool,
) error {
	if version == "" {
		return errors.New("package version is not defined")
	}

	var (
		file *os.File
		err  error
	)

	debugf("getting package file from repository")

	file, err = arch.GetPackageFile(name, version)
	if err != nil {
		return ser.Errorf(
			err,
			"can't get package '%s' from repository", name,
		)
	}

	tracef("changing target repository path to '%s", pathNew)

	arch.SetPath(pathNew)

	debugf("copying package file")

	pathCopied, err := arch.CreatePackageFile(
		path.Base(file.Name()),
		file,
	)
	if err != nil {
		return ser.Errorf(err, "can't copy file to repo")
	}

	err = arch.AddPackage(pathCopied, force)
	if err != nil {
		return ser.Errorf(
			err,
			"can't copy package %s to path %s", name, pathNew,
		)
	}

	return nil
}

func (arch RepositoryArch) RemovePackage(
	name string, version string,
) error {
	if version == "" {
		return errors.New("package version is not defined")
	}

	debugf("getting package file from repository")

	file, err := arch.GetPackageFile(name, version)
	if err != nil {
		return ser.Errorf(
			err,
			"can't get file for package %s", name,
		)
	}

	args := []string{arch.getDatabaseFilepath(), name}

	tracef("executing repo-remove with args: '%v'", args)

	stdout, stderr, err := executil.Run(
		exec.Command("repo-remove", args...),
	)
	if err != nil {
		return ser.Errorf(err, "can't execute repo-remove")
	}

	tracef("repo-remove stdout: %s", stdout)
	tracef("repo-remove stderr: %s", stderr)

	debugf("removing file")

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

	tracef("pacman config: %s", config)

	defer os.RemoveAll(config)

	debugf("executing pacman")

	stdout, stderr, err := executil.Run(
		exec.Command("pacman", []string{
			"--sync", "--info",
			"--config", config,
			"--dbpath", directory, packageName,
		}...),
	)
	if err != nil {
		return "", ser.Errorf(err, "can't execute pacman")
	}

	tracef("pacman stdout: %s", stdout)
	tracef("pacman stderr: %s", stderr)

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
	debugf("generating pacman config")

	debugf("making temporary file")

	config, err := ioutil.TempFile("/tmp/", "repod-pacman-config-")
	if err != nil {
		return "", ser.Errorf(err, "can't make temporary file")
	}

	tracef("temporary file: %s", config.Name())

	debugf("write config content to temporary file")

	content := fmt.Sprintf(formatPacmanConfRepo, arch.getDatabaseName())

	tracef("config file content: '%s'", content)

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
	debugf("getting sync directory")

	debugf("making temporary directory")

	directoryTemp, err := ioutil.TempDir(os.TempDir(), "repod-")
	if err != nil {
		return "", ser.Errorf(
			err,
			"can't make temporary directory",
		)
	}

	tracef("temporary directory: %s", directoryTemp)

	debugf("making sync directory")

	directorySync := filepath.Join(directoryTemp, "/sync")

	err = os.Mkdir(directorySync, 0700)
	if err != nil {
		return "", ser.Errorf(
			err,
			"can't make sync directory %s", directorySync,
		)
	}

	tracef("sync directory: %s", directorySync)

	debugf("symlinking database file to sync directory")

	databaseFile := arch.getDatabaseFilepath()
	if _, err = os.Stat(databaseFile); os.IsNotExist(err) {
		return "", ser.Errorf(
			err,
			"can't stat database file: '%s'",
			databaseFile,
		)
	}

	databaseFileSync := filepath.Join(
		directorySync,
		arch.getDatabaseFilename(),
	)

	tracef(
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
	name string,
	version string,
) (*os.File, error) {
	tracef("finding package '%s' in package dir", name)

	glob := fmt.Sprintf("%s-%s-[a-z0-9_]*.pkg.tar.xz", name, version)

	tracef("package file search pattern: %s", glob)

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
