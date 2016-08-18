package main

import "github.com/gin-gonic/gin"

func runDaemon(
	repoRoot string,
	listenAddress string,
	nucleusAddress string,
	tlsCert string,
) {
	var (
		auth = NucleusAuth{
			Address: nucleusAddress,
		}
		api    = newAPI(repoRoot)
		router = gin.New()
	)

	err := auth.AddCertificateFile(tlsCert)
	if err != nil {
		fatalf("%s", err)
	}

	router.Use(getRouterRecovery(), getRouterLogger())

	v1 := router.Group("/v1/", api.handleAuthentificate)
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
