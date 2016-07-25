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
    -h --help                   Show this help.
    --listen-address=<url>      Address to listen requests to.
                                 [default: :6333]
    --repositories-dir=<path>   Host filesystem directory with repositories.
                                 [default: /srv/http]
`

func main() {
	args, err := docopt.Parse(usage, nil, true, "repod "+version, false)
	if err != nil {
		panic(err)
	}

	api := API{
		RepositoriesDir: args["--repositories-dir"].(string),
	}

	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	v1 := router.Group("/v1/")
	v1.GET("/", api.HandleListRepositories)
	v1.GET("/:repository/", api.HandleListEpoches)
	v1.GET("/:repository/:epoch/", api.HandleListPackages)
	v1.POST("/:repository/:epoch/", api.HandlePackageAdd)
	v1.GET("/:repository/:epoch/:package", api.HandlePackageDescribe)
	v1.DELETE("/:repository/:epoch/:package", api.HandlePackageDelete)
	v1.POST("/:repository/:epoch/:package", api.HandlePackageEdit)

	router.Run(args["--listen-address"].(string))
}
