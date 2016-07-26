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
	repod [--listen-address=<url>] [--repositories-dir=<path>]

Options:
    -h --help                  Show this help.
    --listen-address=<url>     Address to listen requests to.
                                [default: :6333]
    --repositories-dir=<path>  Host filesystem directory with repositories.
                                [default: /srv/http]
`

func main() {
	args, err := docopt.Parse(usage, nil, true, "repod "+version, false)
	if err != nil {
		panic(err)
	}

	var (
		repositoriesDir = args["--repositories-dir"].(string)
		listenAddress   = args["--listen-address"].(string)

		api    = newAPI(repositoriesDir)
		router = gin.New()
	)

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(api.DetectRepositoryOS)

	v1 := router.Group("/v1/")
	v1.GET("/", api.HandleListRepositories)
	v1.GET("/:repo/", api.HandleListEpoches)
	v1.GET(urlListPackages, api.HandleListPackages)
	v1.PUT(urlPackage, api.HandleAddPackage)
	v1.GET(urlPackage, api.HandleDescribePackage)
	v1.DELETE(urlPackage, api.HandleDeletePackage)
	v1.POST(urlPackage, api.HandleEditPackage)

	router.Run(listenAddress)
}
