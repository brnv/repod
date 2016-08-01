package main

import (
	docopt "github.com/docopt/docopt-go"
	"github.com/gin-gonic/gin"
)

var version = "1.0"

var usage = `repod - daemon to manage packages repository stored on host.

Required. TODO.

Usage:
	repod -h | --help
	repod [--listen=<url>] [--repos-dir=<path>]

Options:
    -h --help           Show this help.
    --listen=<url>      Address to listen requests to.
                         [default: :6333]
    --repos-dir=<path>  Specify directory where repos stored.
                         [default: /srv/http]
`

func main() {
	args, err := docopt.Parse(usage, nil, true, "repod "+version, false)
	if err != nil {
		panic(err)
	}

	var (
		repositoriesDir = args["--repos-dir"].(string)
		listenAddress   = args["--listen"].(string)

		api    = newAPI(repositoriesDir)
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
			"GET", "/:repo/",
			api.handleListEpoches,
		)
		v1.Handle(
			"GET", urlListPackages,
			api.handleListPackages,
		)
		v1.Handle(
			"POST", urlPackage,
			api.handleAddPackage,
		)
		v1.Handle(
			"GET", urlPackage,
			api.handleDescribePackage,
		)
		v1.Handle(
			"DELETE", urlPackage,
			api.handleRemovePackage,
		)
		v1.Handle(
			"PATCH", urlPackage,
			api.handleEditPackage,
		)
	}

	router.Run(listenAddress)
}
