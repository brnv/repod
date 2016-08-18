package main

import (
	"fmt"
	"os"
	"strings"

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
    repod [options] --listen <address> --nucleus <address> --tls-cert <path>
    repod [options] -L
    repod [options] -L <repo>
    repod [options] -L <repo> <epoch> <db> <arch>
    repod [options] (-A|-S|-E|-R) <repo> <epoch> <db> <arch> <package>

Options:
    --root <path>             Directory where repositories stored
                               [default: /srv/http].
    --listen <address>        Listen specified IP and port.
    --nucleus <address>       Nucleus server address.
    --tls-cert <path>         Path to nucleus ssl certificate file.
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

		repo, _        = args["<repo>"].(string)
		epoch, _       = args["<epoch>"].(string)
		db, _          = args["<db>"].(string)
		arch, _        = args["<arch>"].(string)
		packageName, _ = args["<package>"].(string)
		packageFile, _ = args["--file"].(string)

		epochToChange, _ = args["--change-epoch"].(string)

		listenAddress, _  = args["--listen"].(string)
		nucleusAddress, _ = args["--nucleus"].(string)
		tlsCert, _        = args["--tls-cert"].(string)

		err        error
		output     string
		repository Repository
	)

	if listenAddress != "" {
		runDaemon(repoRoot, listenAddress, nucleusAddress, tlsCert)
	}

	if args["--list"].(bool) && repo == "" {
		repos := []string{}
		repos, err = listRepositories(repoRoot)
		output = strings.Join(repos, "\n")
	} else {
		repository, err = getRepository(repo, repoRoot, epoch, db, arch)
		if err != nil {
			fatalf("%s", err)
		}
	}

	switch {
	case args["--list"].(bool):
		if repo != "" {
			output, err = listEpoches(repoRoot, repository)
		}

		if repo != "" && arch != "" {
			output, err = listPackages(repoRoot, repository)
		}

	case args["--add"].(bool):
		output, err = addPackage(
			repoRoot, repository, packageName, packageFile,
		)

	case args["--show"].(bool):
		output, err = describePackage(
			repoRoot, repository, packageName,
		)

	case args["--edit"].(bool):
		output, err = editPackage(
			repoRoot, repository, packageName,
			packageFile, epochToChange,
		)

	case args["--remove"].(bool):
		err = repository.RemovePackage(packageName)

	}

	if err != nil {
		fatalf("%s", err)
	}

	fmt.Print(output)
}
