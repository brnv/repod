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

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(api.detectRepositoryOS)

	v1 := router.Group("/v1/")
	v1.GET("/", api.handleListRepositories)
	v1.GET("/:repo/", api.handleListEpoches)
	v1.GET(urlListPackages, api.handleListPackages)
	v1.PUT(urlPackage, api.handleAddPackage)
	v1.GET(urlPackage, api.handleDescribePackage)
	v1.DELETE(urlPackage, api.handleRemovePackage)
	v1.POST(urlPackage, api.handleEditPackage)

	router.Run(listenAddress)
}
