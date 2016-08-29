package main

import "github.com/gin-gonic/gin"

const (
	urlList    = "/list"
	urlAdd     = "/add"
	urlPackage = "/package/:name"
)

func runDaemon(
	repoRoot string,
	listenAddress string,
	nucleusAddress string,
	tlsCert string,
) {
	var (
		nucleusAuth = NucleusAuth{}
		router      = gin.New()
	)

	api := newAPI(repoRoot)

	if nucleusAddress != "" {
		api.authNeed = true
		nucleusAuth.Address = nucleusAddress
	}

	if tlsCert != "" {
		err := nucleusAuth.AddCertificateFile(tlsCert)
		if err != nil {
			fatalf("%s", err)
		}
	}

	router.Use(getRouterRecovery(), getRouterLogger())

	v1 := router.Group("/v1/", api.handleAuthentificate)
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

	router.Run(listenAddress)
}
