package main

import "github.com/gin-gonic/gin"

const (
	urlEpoches       = "/:repo"
	urlDatabases     = urlEpoches + "/:epoch"
	urlArchitectures = urlDatabases + "/:db"
	urlPackages      = urlArchitectures + "/:arch"
	urlPackage       = urlPackages + "/:package"
)

func runDaemon(
	repoRoot string,
	listenAddress string,
	nucleusAddress string,
	tlsCert string,
) {
	var (
		nucleusAuth = NucleusAuth{}
		api         = newAPI(repoRoot)
		router      = gin.New()
	)

	if nucleusAddress != "" {
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
			"GET", urlEpoches,
			api.handleListEpoches,
		)
		v1.Handle(
			"GET", urlPackages,
			api.handleListPackages,
		)
		v1.Handle(
			"POST", urlPackages,
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
