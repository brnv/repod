package main

import (
	"git.rn/devops/nucleus-go"
	"github.com/gin-gonic/gin"
	"github.com/reconquest/ser-go"
)

func runServer(
	repoRoot string, listenAddress string,
	nucleusAddress string, tlsCert string,
) error {
	debugf("running server")

	var (
		router = gin.New()
		api    = &API{root: repoRoot}
		err    error
	)

	tracef("api: %#v", api)

	if nucleusAddress != "" {
		debugf("nucleus authorization required")

		api.authEnabled = true
	}

	if tlsCert != "" {
		debugf("using given tls certificate")

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

	router.Use(getRouterRecovery, getRouterLogger)

	v1 := router.Group(
		"/v1/",
		api.handleAuthentificate,
		api.prepareResponse,
		api.readRequestParams,
	)
	{
		v1.Handle(
			"GET", "/",
			api.handleListRepositories,
		)
		v1.Handle(
			"GET", "/list",
			api.handleListPackages,
		)
		v1.Handle(
			"POST", "/add",
			api.handleAddPackage,
		)
		v1.Handle(
			"GET", "/package/:name",
			api.handleDescribePackage,
		)
		v1.Handle(
			"DELETE", "/package/:name",
			api.handleRemovePackage,
		)
		v1.Handle(
			"POST", "/package/:name",
			api.handleCopyPackage,
		)
	}

	return router.Run(listenAddress)
}
