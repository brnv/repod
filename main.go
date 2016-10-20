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

	exit = os.Exit
)

var usage = `repod, packages repository manager

Usage:
    repod -h | --help
    repod [options] --listen <address> [--nucleus <address> --tls-cert <path>]
    repod [options] -Q [<path>]
    repod [options] -A <path> -f <path> [--force]
    repod [options] (-C|-S|-R) <path> <package>

Options:
    -d --root <path>        Repositories directory [default: /srv/http].
    -l --listen <address>   Listen specified IP and port.
    -n --nucleus <address>  Nucleus server address.
    -c --tls-cert <path>    Path to nucleus ssl certificate file.
    -C --copy               Copy package to another db path.
      -t --copy-to <db>     Specify db path to copy package to.
    -S --show               Describe package.
    -R --remove             Remove package.
    -A --add                Add package.
      -f --file <path>      Specify file to be upload to repository.
      --force               Force add.
    -Q --query              List packages and repositories.
      <path>                Target repository path.
      <package>             Target package to manipulate with.
    -s --system <type>      Specify repository system [default: autodetect]
    --debug                 Show runtime debug information.
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

		forceAdd, _ = args["--force"].(bool)

		pathNew, _ = args["--copy-to"].(string)

		listenAddress, _  = args["--listen"].(string)
		nucleusAddress, _ = args["--nucleus"].(string)
		tlsCert, _        = args["--tls-cert"].(string)

		err        error
		output     string
		repository Repository

		system = args["--system"].(string)
	)

	if args["--debug"].(bool) {
		logger.SetLevel(lorg.LevelDebug)
	}

	if listenAddress != "" {
		fatalln(
			runServer(root, listenAddress, nucleusAddress, tlsCert),
		)

		return
	}

	if args["--query"].(bool) && path == "" {
		var repositories []string

		repositories, err = listRepositories(root)

		if len(repositories) > 0 {
			fmt.Println(strings.Join(repositories, "\n"))
		}

		return
	}

	repository, err = getRepository(root, path, system)
	if err != nil {
		fatalln(err)
	}

	switch {

	case args["--query"].(bool):
		var packages []string

		packages, err = repository.ListPackages()

		if len(packages) > 0 {
			output = strings.Join(packages, "\n")
		}

	case args["--add"].(bool):
		err = addPackage(repository, packagePath, forceAdd)

	case args["--show"].(bool):
		output, err = repository.DescribePackage(packageName)

	case args["--copy"].(bool):
		err = repository.CopyPackage(packageName, pathNew)

	case args["--remove"].(bool):
		err = repository.RemovePackage(packageName)

	}

	if err != nil {
		fatalln(err)
	}

	if len(output) > 0 {
		fmt.Println(output)
	}
}
