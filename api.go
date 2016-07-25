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
	RepositoriesDir string
	RepositoryOS    int
}

type Response struct {
	Success bool                `json:"success"`
	Error   string              `json:"error"`
	Data    map[string][]string `json:"data"`

	status int
}

const (
	urlListPackages = "/:repo/:epoch/:db/:arch"
	urlPackage      = urlListPackages + "/:package"

	dataKeyRepositories = "repositories"
	dataKeyEpoches      = "epoches"
	dataKeyPackages     = "packages"

	formatRepositoryPath = "%s/%s/"

	osArchLinux = 1
	osUbuntu    = 2

	formPackageFile = "package_file"
)

var (
	defaultResponse = Response{
		Data:    make(map[string][]string),
		status:  http.StatusOK,
		Success: true,
	}
)

func (api *API) DetectRepositoryOS(context *gin.Context) {
	repository := context.Param("repo")

	if strings.HasPrefix(repository, "arch") {
		api.RepositoryOS = osArchLinux
	}

	if strings.HasPrefix(repository, "ubuntu") {
		api.RepositoryOS = osUbuntu
	}
}

func (api API) HandleListRepositories(context *gin.Context) {
	response := defaultResponse

	repositories, err := ioutil.ReadDir(api.RepositoriesDir)
	if err != nil {
		api.sendResponse(context, api.getErrorResponse(err))
		return
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
		repository    = context.Param("repo")
		repositoryDir = fmt.Sprintf(
			formatRepositoryPath,
			api.RepositoriesDir,
			repository,
		)
	)

	epoches, err := ioutil.ReadDir(repositoryDir)
	if err != nil {
		api.sendResponse(context, api.getErrorResponse(err))
		return
	}

	for _, epoch := range epoches {
		response.Data[dataKeyEpoches] = append(
			response.Data[dataKeyEpoches], epoch.Name(),
		)
	}

	api.sendResponse(context, response)
}

func (api *API) HandleListPackages(context *gin.Context) {
	var (
		err      error
		response = defaultResponse
	)

	repository, err := api.getRepository(context)
	if err != nil {
		api.sendResponse(context, api.getErrorResponse(err))
		return
	}

	response.Data[dataKeyPackages], err = repository.ListPackages()
	if err != nil {
		api.sendResponse(context, api.getErrorResponse(err))
		return
	}

	api.sendResponse(context, response)
}

func (api *API) HandleAddPackage(context *gin.Context) {
	var (
		err      error
		response = defaultResponse
	)

	repository, err := api.getRepository(context)
	if err != nil {
		api.sendResponse(context, api.getErrorResponse(err))
		return
	}

	packageFile, packageFileHeader, err :=
		context.Request.FormFile(formPackageFile)
	if err != nil {
		api.sendResponse(
			context,
			api.getErrorResponse(
				errors.New("can't read package file form file"),
			),
		)
		return
	}

	err = repository.AddPackage(
		packageFileHeader.Filename,
		packageFile,
	)
	if err != nil {
		api.sendResponse(context, api.getErrorResponse(err))
		return
	}

	api.sendResponse(context, response)
}

func (api *API) HandleDeletePackage(context *gin.Context) {
	api.sendResponse(context, defaultResponse)
}

func (api *API) HandleEditPackage(context *gin.Context) {
	api.sendResponse(context, defaultResponse)
}

func (api *API) HandleDescribePackage(context *gin.Context) {
	api.sendResponse(context, defaultResponse)
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

func (api *API) getRepository(context *gin.Context) (Repository, error) {
	repositoryDir := fmt.Sprintf(
		formatRepositoryPath,
		api.RepositoriesDir,
		context.Param("repo"),
	)

	if api.RepositoryOS == osArchLinux {
		return &RepositoryArch{
			Path:         repositoryDir,
			Epoch:        context.Param("epoch"),
			Database:     context.Param("db"),
			Architecture: context.Param("arch"),
		}, nil
	}

	return nil, fmt.Errorf("can't get repository from request")
}
