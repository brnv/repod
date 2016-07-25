package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kovetskiy/toml"
)

type API struct {
	RepositoriesDir string
}

type Response struct {
	Success bool                `json:"success"`
	Error   string              `json:"error"`
	Data    map[string][]string `json:"data"`

	status int
}

const (
	dataKeyRepositories = "repositories"
	dataKeyEpoches      = "epoches"

	formatRepositoryPath = "%s/%s/"
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

	repositories, err := ioutil.ReadDir(api.RepositoriesDir)
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

func (api *API) HandleListEpoches(context *gin.Context) {
	var (
		response      = defaultResponse
		repository    = context.Param("repository")
		repositoryDir = fmt.Sprintf(
			formatRepositoryPath,
			api.RepositoriesDir,
			repository,
		)
	)

	epoches, err := ioutil.ReadDir(repositoryDir)
	if err != nil {
		response = api.getErrorResponse(err)
	}

	for _, epoch := range epoches {
		response.Data[dataKeyEpoches] = append(
			response.Data[dataKeyEpoches], epoch.Name(),
		)
	}

	api.sendResponse(context, response)
}

func (api *API) HandleListPackages(context *gin.Context) {

}

func (api *API) HandleAddPackage(context *gin.Context) {

}

func (api *API) HandleDeletePackage(context *gin.Context) {

}

func (api *API) HandleEditPackage(context *gin.Context) {

}

func (api *API) HandleDescribePackage(context *gin.Context) {

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
