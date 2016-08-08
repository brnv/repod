package main

import (
	"github.com/gin-gonic/gin"
)

type API struct {
	RepositoryDir string
}

func (api API) HandleListRepositories(context *gin.Context) {

}

func (api *API) HandleListEpochs(context *gin.Context) {

}

func (api *API) HandleListPackages(context *gin.Context) {

}

func (api *API) HandlePackageAdd(context *gin.Context) {

}

func (api *API) HandlePackageDelete(context *gin.Context) {

}

func (api *API) HandlePackageEdit(context *gin.Context) {

}

func (api *API) HandlePackageDescribe(context *gin.Context) {

}
