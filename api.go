package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kovetskiy/toml"
)

type API struct {
	repositoriesDir string
	repositoryOS    string

	defaultResponse APIResponse
}

type APIResponse struct {
	Success bool                `json:"success"`
	Error   string              `json:"error"`
	Data    map[string][]string `json:"data"`

	status int
}

const (
	urlListPackages = "/:repo/:epoch/:db/:arch"
	urlPackage      = urlListPackages + "/:package"

	sliceKeyRepositories = "repositories"
	sliceKeyEpoches      = "epoches"
	sliceKeyPackages     = "packages"

	osArchLinux = "arch"
	osUbuntu    = "ubuntu"

	postFormPackageFile = "package_file"
)

func newAPI(repositoriesDir string) *API {
	return &API{
		repositoriesDir: repositoriesDir,
		defaultResponse: APIResponse{
			Data:    make(map[string][]string),
			status:  http.StatusOK,
			Success: true,
		},
	}
}

func (api *API) detectRepositoryOS(context *gin.Context) {
	repository := context.Param("repo")

	if strings.HasPrefix(repository, "arch") {
		api.repositoryOS = osArchLinux
	}

	if strings.HasPrefix(repository, "ubuntu") {
		api.repositoryOS = osUbuntu
	}
}

func (api API) handleListRepositories(context *gin.Context) {
	response := api.defaultResponse

	repositories, err := ioutil.ReadDir(api.repositoriesDir)
	if err != nil {
		api.sendResponse(context, api.getErrorResponse(err))
		return
	}

	for _, repository := range repositories {
		response.Data[sliceKeyRepositories] = append(
			response.Data[sliceKeyRepositories], repository.Name(),
		)
	}

	api.sendResponse(context, response)
}

func (api *API) handleListEpoches(context *gin.Context) {
	var (
		response      = api.defaultResponse
		repositoryDir = api.repositoriesDir + "/" + context.Param("repo")
	)

	epoches, err := ioutil.ReadDir(repositoryDir)
	if err != nil {
		api.sendResponse(context, api.getErrorResponse(err))
		return
	}

	for _, epoch := range epoches {
		response.Data[sliceKeyEpoches] = append(
			response.Data[sliceKeyEpoches], epoch.Name(),
		)
	}

	api.sendResponse(context, response)
}

func (api *API) handleListPackages(context *gin.Context) {
	var (
		err      error
		response = api.defaultResponse
	)

	repository, err := api.newRepository(context)
	if err != nil {
		api.sendResponse(context, api.getErrorResponse(err))
		return
	}

	response.Data[sliceKeyPackages], err = repository.ListPackages()
	if err != nil {
		api.sendResponse(context, api.getErrorResponse(err))
		return
	}

	api.sendResponse(context, response)
}

func (api *API) handleAddPackage(context *gin.Context) {
	var (
		err      error
		response = api.defaultResponse
		request  = context.Request
	)

	repository, err := api.newRepository(context)
	if err != nil {
		api.sendResponse(context, api.getErrorResponse(err))
		return
	}

	file, fileHeader, err := request.FormFile(postFormPackageFile)
	if err != nil {
		api.sendResponse(
			context,
			api.getErrorResponse(
				errors.New("can't read package file form file"),
			),
		)
		return
	}

	err = repository.AddPackage(fileHeader.Filename, file)
	if err != nil {
		api.sendResponse(context, api.getErrorResponse(err))
		return
	}

	api.sendResponse(context, response)
}

func (api *API) handleRemovePackage(context *gin.Context) {
	var (
		err         error
		response    = api.defaultResponse
		packageName = context.Param("package")
	)

	repository, err := api.newRepository(context)
	if err != nil {
		api.sendResponse(context, api.getErrorResponse(err))
		return
	}

	err = repository.RemovePackage(packageName)
	if err != nil {
		api.sendResponse(context, api.getErrorResponse(err))
		return
	}

	api.sendResponse(context, response)
}

func (api *API) handleEditPackage(context *gin.Context) {
	api.sendResponse(context, api.defaultResponse)
}

func (api *API) handleDescribePackage(context *gin.Context) {
	api.sendResponse(context, api.defaultResponse)
}

func (api *API) sendResponse(
	context *gin.Context, response APIResponse,
) {
	err := toml.NewEncoder(context.Writer).Encode(response)
	if err != nil {
		log.Printf("can't send response %#v", response)
		context.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	context.Writer.WriteHeader(response.status)
}

func (api *API) getErrorResponse(err error) APIResponse {
	return APIResponse{
		Success: false,
		Error:   err.Error(),
		status:  http.StatusInternalServerError,
	}
}

func (api *API) newRepository(context *gin.Context) (Repository, error) {
	switch api.repositoryOS {
	case osArchLinux:
		return &RepositoryArch{
			path:         api.repositoriesDir + "/" + context.Param("repo"),
			epoch:        context.Param("epoch"),
			database:     context.Param("db"),
			architecture: context.Param("arch"),
		}, nil
	}

	return nil, fmt.Errorf("can't get repository from request")
}
