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

	exit = os.Exit
)

var usage = `repod, packages repository manager

Usage:
  repod -h | --help
  repod [options] --listen <address> [--nucleus <address> --tls-cert <path>]
  repod [options] -Q [<path>]
  repod [options] -A <path> -f <path>
  repod [options] (-E|-S|-R) <path> <package> [-f <path>]

Options:
  -d --root <path>        Repositories directory [default: /srv/http].
  -l --listen <address>   Listen specified IP and port.
  -n --nucleus <address>  Nucleus server address.
  -c --tls-cert <path>    Path to nucleus ssl certificate file.
  -E --edit               Edit package file, database or description.
    -t --copy-to <path>   Specify database to copy package to.
  -S --show               Describe package.
  -R --remove             Remove package.
  -A --add                Add package.
    -f --file <path>      Specify file to be upload to repository.
  -Q --query              List packages and repositories.
    <path>                Target repository path.
    <package>             Target package to manipulate with.
  -s --system <type>      Specify repository system [default: autodetect]
  -h --help               Show this help.
`

func main() {
	args := godocs.MustParse(usage, version, godocs.UsePager)

	logger.SetIndentLines(true)

	var (
		root    = args["--root"].(string)
		path, _ = args["<path>"].(string)

		packageName, _ = args["<package>"].(string)
		packagePath, _ = args["--file"].(string)

		rootNew, _ = args["--copy-to"].(string)

		listenAddress, _  = args["--listen"].(string)
		nucleusAddress, _ = args["--nucleus"].(string)
		tlsCert, _        = args["--tls-cert"].(string)

		modeQuery, _         = args["--query"].(bool)
		modeListRepositories = modeQuery && path == ""

		err        error
		output     string
		repository Repository

		system = args["--system"].(string)
	)

	if listenAddress != "" {
		runDaemon(root, listenAddress, nucleusAddress, tlsCert)
	}

	repository, err = getRepository(root, path, system)
	if err != nil && !modeListRepositories {
		fatalln(err)
	}

	switch {
	case modeListRepositories:
		output, err = listRepositories(root)

	case args["--query"].(bool):
		if path != "" {
			output, err = listPackages(repository)
		}

	case args["--add"].(bool):
		output, err = addPackage(repository, packagePath)

	case args["--show"].(bool):
		output, err = describePackage(
			repository, packageName,
		)

	case args["--edit"].(bool):
		output, err = editPackage(
			repository, packageName, packagePath, rootNew,
		)

	case args["--remove"].(bool):
		output, err = repository.RemovePackage(packageName)

	}

	if err != nil {
		fatalf("%s", err)
	}

	if len(output) > 0 {
		fmt.Println(output)
	}
}
