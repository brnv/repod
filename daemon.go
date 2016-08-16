package main

import "github.com/gin-gonic/gin"

func runDaemon(repoRoot string, listenAddress string) {
	api := newAPI(repoRoot)

	router := gin.New()

	router.Use(getRouterRecovery(), getRouterLogger())

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
