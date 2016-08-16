package main

import (
	"fmt"
	"os"

	"github.com/kovetskiy/godocs"
	"github.com/kovetskiy/lorg"
)

var (
	logger  = lorg.NewLog()
	version = "1"
	exit    = os.Exit
)

const (
	urlListEpoches       = "/:repo"
	urlListDatabases     = urlListEpoches + "/:epoch"
	urlListArchitectures = urlListDatabases + "/:db"
	urlListPackages      = urlListArchitectures + "/:arch"
	urlManipulatePackage = urlListPackages + "/:package"
)

var usage = `repod, packages repository manager

Usage:
    repod -h | --help
    repod [options] --listen <address>
    repod [options] -L
    repod [options] -L <repo>
    repod [options] -L <repo> <epoch> <db> <arch>
    repod [options] (-A|-S|-E|-R) <repo> <epoch> <db> <arch> <package>

Options:
    --root <path>             Directory where repositories stored
                               [default: /srv/http].
    --listen <address>        Listen specified IP and port.
    -L --list                 List packages, epoches or repositories.
    -A --add                  Add package.
    -S --show                 Describe package.
    -E --edit                 Edit package file, epoch or description.
    -R --remove               Remove package.
      <repo>                  Target repository name.
      <epoch>                 Target repository epoch.
      <db>                    Target repository database.
      <arch>                  Target repository architecture.
      <package>               Target package to manipulate with.
      --file <path>           Specify file to be upload to repository.
      --change-epoch <epoch>  Specify epoch to copy package to.
    -h --help                 Show this help.
`

func main() {
	args := godocs.MustParse(usage, version, godocs.UsePager)

	var (
		repoRoot = args["--root"].(string)

		actionList   = args["--list"].(bool)
		actionAdd    = args["--add"].(bool)
		actionShow   = args["--show"].(bool)
		actionEdit   = args["--edit"].(bool)
		actionRemove = args["--remove"].(bool)

		repo, _        = args["<repo>"].(string)
		epoch, _       = args["<epoch>"].(string)
		db, _          = args["<db>"].(string)
		arch, _        = args["<arch>"].(string)
		packageName, _ = args["<package>"].(string)
		packageFile, _ = args["--file"].(string)

		epochToChange, _ = args["--change-epoch"].(string)

		listenAddress, _ = args["--listen"].(string)

		err    error
		output string
	)

	if listenAddress != "" {
		runDaemon(repoRoot, listenAddress)
	}

	if actionList && repo == "" {
		output, err = getListRepositoriesOutput(repoRoot)
		if err != nil {
			fatalf("%s", err)
		}
	}

	repository, err := getRepository(repo, repoRoot, epoch, db, arch)
	if err != nil {
		fatalf("%s", err)
	}

	if actionList && repo != "" {
		output, err = getListEpochesOutput(repoRoot, repository)
	}

	if actionList && repo != "" && arch != "" {
		output, err = getListPackagesOutput(repoRoot, repository)
	}

	if actionAdd {
		output, err = getAddPackageOutput(
			repoRoot, repository, packageName, packageFile,
		)
	}

	if actionShow {
		output, err = getDescribePackageOutput(
			repoRoot, repository, packageName,
		)
	}

	if actionEdit {
		output, err = getEditPackageOutput(
			repoRoot, repository, packageName,
			packageFile, epochToChange,
		)
	}

	if actionRemove {
		err = repository.RemovePackage(packageName)
	}

	if err != nil {
		fatalf("%s", err)
	}

	fmt.Print(output)
}
