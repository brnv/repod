package main

import (
	"fmt"
	"os"
	"path/filepath"

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

var usage = `repod - daemon to manage packages repository stored on host.

Usage:
    repod -h | --help
    repod [options] --listen <address>
    repod [options] -L
    repod [options] -L <repo>
    repod [options] -L <repo> <epoch> <db> <arch>
    repod [options] (-A|-S|-E|-R) <repo> <epoch> <db> <arch> <package>

Options:
    --root=<path>            Specify directory where repos stored.
                              [default: /srv/http]
    --listen <address>       Address to listen requests to.
    -L --list                List repositories or repository packages.
    -A --add                 Add package to specified repository.
    -S --show                Show package information.
    -E --edit                Edit package in specified repository.
    -R --remove              Remove package from specified repository.
      <repo>                 Specify repository name.
      <epoch>                Specify repository epoch.
      <db>                   Specify repository db.
      <arch>                 Specify repository architecture.
      <package>              Specify package to manipulate with.
     --file=<path>           Package file to add or edit in repository.
     --change-epoch=<epoch>  Package file to add or edit in repository.
    -h --help                Show this help.
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

	osType := detectRepositoryOS(repo)

	repoPath := filepath.Join(repoRoot, repo)

	repository, err := getRepository(osType, repoPath, epoch, db, arch)
	if err != nil {
		fatalf("%s", err)
	}

	if actionList && repo != "" {
		output, err = getListEpochesOutput(repoRoot, repository)
		if err != nil {
			fatalf("%s", err)
		}
	}

	if actionList && repo != "" && arch != "" {
		output, err = getListPackagesOutput(repoRoot, repository)
		if err != nil {
			fatalf("%s", err)
		}
	}

	if actionAdd {
		output, err = getAddPackageOutput(
			repoRoot, repository, packageName, packageFile,
		)
		if err != nil {
			fatalf("%s", err)
		}
	}

	if actionShow {
		output, err = getDescribePackageOutput(
			repoRoot, repository, packageName,
		)
		if err != nil {
			fatalf("%s", err)
		}
	}

	if actionEdit {
		output, err = getEditPackageOutput(
			repoRoot, repository, packageName,
			packageFile, epochToChange,
		)
		if err != nil {
			fatalf("%s", err)
		}
	}

	if actionRemove {
		err = repository.RemovePackage(packageName)
		if err != nil {
			fatalf("%s", err)
		}
	}

	fmt.Print(output)
}
