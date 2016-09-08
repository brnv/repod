package main

import (
	"git.rn/devops/nucleus-go"
	"github.com/gin-gonic/gin"
	ser "github.com/reconquest/ser-go"
)

const (
	urlList    = "/list"
	urlAdd     = "/add"
	urlPackage = "/package/:name"
)

func runServer(
	repoRoot string, listenAddress string,
	nucleusAddress string, tlsCert string,
) error {
	tracef("running server")

	var (
		router = gin.New()
		api    = &API{root: repoRoot}
		err    error
	)

	debugf("api: %#v", api)

	if nucleusAddress != "" {
		tracef("nucleus authorization required")

		api.nucleusAuth = true
	}

	if tlsCert != "" {
		tracef("using given tls certificate")

		err = nucleus.AddCertificateFile(tlsCert)
		if err != nil {
			return ser.Errorf(
				err,
				"can't add certificate to nucleus",
			)
		}

		nucleus.SetAddress(nucleusAddress)
		nucleus.SetUserAgent("repod/" + version)
	}

	router.Use(getRouterRecovery(), getRouterLogger())

	v1 := router.Group(
		"/v1/",
		api.handleAuthentificate,
		api.prepareResponse,
		api.detectRepository,
	)

	{
		v1.Handle(
			"GET", "/",
			api.handleListRepositories,
		)
		v1.Handle(
			"GET", urlList,
			api.handleListPackages,
		)
		v1.Handle(
			"POST", urlAdd,
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

	return router.Run(listenAddress)
}
