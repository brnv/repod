package main

import (
	docopt "github.com/docopt/docopt-go"
	"github.com/gin-gonic/gin"
)

var version = "1.0"

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
	repod [--listen=<url>] [--root=<path>]

Options:
    -h --help       Show this help.
    --listen=<url>  Address to listen requests to.
                     [default: :6333]
    --root=<path>   Specify directory where repos stored.
                     [default: /srv/http]
`

func main() {
	args, err := docopt.Parse(usage, nil, true, "repod "+version, false)
	if err != nil {
		panic(err)
	}

	var (
		repoRoot      = args["--root"].(string)
		listenAddress = args["--listen"].(string)

		api    = newAPI(repoRoot)
		router = gin.New()
	)

	router.Use(gin.Logger(), gin.Recovery(), api.detectRepositoryOS)

	v1 := router.Group("/v1/")
	{
		v1.Handle(
			"GET", "/",
			api.handleListRepositories,
		)
		v1.Handle(
			"GET", urlListEpoches,
			api.handleListEpoches,
		)
		v1.Handle(
			"GET", urlListPackages,
			api.handleListPackages,
		)
		v1.Handle(
			"POST", urlManipulatePackage,
			api.handleAddPackage,
		)
		v1.Handle(
			"GET", urlManipulatePackage,
			api.handleDescribePackage,
		)
		v1.Handle(
			"DELETE", urlManipulatePackage,
			api.handleRemovePackage,
		)
		v1.Handle(
			"PATCH", urlManipulatePackage,
			api.handleEditPackage,
		)
	}

	router.Run(listenAddress)
}
