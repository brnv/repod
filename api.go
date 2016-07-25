package main

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kovetskiy/toml"
)

type API struct {
	RepositoryDir string
}

type Response struct {
	Success bool                `json:"success"`
	Error   string              `json:"error"`
	Data    map[string][]string `json:"data"`

	status int
}

const (
	dataKeyRepositories = "repositories"
)

var (
	defaultResponse = Response{
		Data:    make(map[string][]string),
		status:  http.StatusOK,
		Success: true,
	}
)

func (api API) HandleListRepositories(context *gin.Context) {
	response := defaultResponse

	repositories, err := ioutil.ReadDir(api.RepositoryDir)
	if err != nil {
		response = api.getErrorResponse(err)
	}

	for _, repository := range repositories {
		response.Data[dataKeyRepositories] = append(
			response.Data[dataKeyRepositories], repository.Name(),
		)
	}

	api.sendResponse(context, response)
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

func (api *API) sendResponse(
	context *gin.Context, response Response,
) {
	err := toml.NewEncoder(context.Writer).Encode(response)
	if err != nil {
		log.Printf("can't send response %#v", response)
		context.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	context.Writer.WriteHeader(response.status)
}

func (api *API) getErrorResponse(err error) Response {
	return Response{
		Success: false,
		Error:   err.Error(),
		status:  http.StatusInternalServerError,
	}
}