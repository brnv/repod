package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kovetskiy/toml"
)

type API struct {
	repositoriesDir string
	repositoryOS    string
}

type APIResponse struct {
	Success bool                `json:"success"`
	Error   string              `json:"error"`
	Data    map[string][]string `json:"data"`

	status int
}

const (
	urlListEpoches       = "/:repo"
	urlListDatabases     = urlListEpoches + "/:epoch"
	urlListArchitectures = urlListDatabases + "/:db"
	urlListPackages      = urlListArchitectures + "/:arch"
	urlManipulatePackage = urlListPackages + "/:package"

	sliceKeyRepositories = "repositories"
	sliceKeyEpoches      = "epoches"
	sliceKeyPackages     = "packages"
	sliceKeyPackage      = "package"

	osArchLinux = "arch"
	osUbuntu    = "ubuntu"

	postFormPackageFile = "package_file"
)

func newAPI(repositoriesDir string) *API {
	return &API{
		repositoriesDir: repositoriesDir,
	}
}

func newAPIResponse() APIResponse {
	return APIResponse{
		Data:    make(map[string][]string),
		status:  http.StatusOK,
		Success: true,
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
	response := newAPIResponse()

	repositories, err := ioutil.ReadDir(api.repositoriesDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = errors.New("unknown repo")
		}
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
		response      = newAPIResponse()
		repositoryDir = api.repositoriesDir + "/" + context.Param("repo")
	)

	epoches, err := ioutil.ReadDir(repositoryDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = errors.New("unknown repo")
		}
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
		response = newAPIResponse()
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
		err         error
		response    = newAPIResponse()
		request     = context.Request
		packageName = context.Param("package")
	)

	repository, err := api.newRepository(context)
	if err != nil {
		api.sendResponse(context, api.getErrorResponse(err))
		return
	}

	file, _, err := request.FormFile(postFormPackageFile)
	if err != nil {
		api.sendResponse(
			context,
			api.getErrorResponse(
				fmt.Errorf(
					"can't read package file form file: %s", err.Error(),
				),
			),
		)
		return
	}

	err = repository.AddPackage(packageName, file)
	if err != nil {
		api.sendResponse(context, api.getErrorResponse(err))
		return
	}

	api.sendResponse(context, response)
}

func (api *API) handleRemovePackage(context *gin.Context) {
	var (
		err         error
		response    = newAPIResponse()
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
	err := context.Request.ParseForm()
	if err != nil {
		api.sendResponse(context, api.getErrorResponse(err))
		return
	}

	if api.shouldChangePackageEpoch(context) {
		api.handleChangePackageEpoch(context)
		return
	}

	api.handleAddPackage(context)

}

func (api *API) handleChangePackageEpoch(context *gin.Context) {
	var (
		packageName = context.Param("package")
		newEpoch    = context.Request.Form.Get("new_epoch")
		response    = newAPIResponse()
	)

	repository, err := api.newRepository(context)
	if err != nil {
		api.sendResponse(context, api.getErrorResponse(err))
		return
	}

	err = repository.ChangePackageEpoch(packageName, newEpoch)
	if err != nil {
		api.sendResponse(context, api.getErrorResponse(err))
		return
	}

	api.sendResponse(context, response)
}

func (api *API) handleDescribePackage(context *gin.Context) {
	var (
		err         error
		response    = newAPIResponse()
		packageName = context.Param("package")
	)

	repository, err := api.newRepository(context)
	if err != nil {
		api.sendResponse(context, api.getErrorResponse(err))
		return
	}

	response.Data[sliceKeyPackage], err =
		repository.DescribePackage(packageName)
	if err != nil {
		api.sendResponse(context, api.getErrorResponse(err))
		return
	}

	api.sendResponse(context, response)
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
	var (
		err          error
		repoPath     = api.repositoriesDir + "/" + context.Param("repo")
		epoch        = context.Param("epoch")
		database     = context.Param("db")
		architecture = context.Param("arch")
	)

	err = api.ensureRepositoryPaths(repoPath, epoch, database, architecture)
	if err != nil {
		return nil, err
	}

	switch api.repositoryOS {
	case osArchLinux:
		return &RepositoryArch{
			path:         repoPath,
			epoch:        epoch,
			database:     database,
			architecture: architecture,
		}, nil
	}

	return nil, fmt.Errorf("can't detect repository from request")
}

func (api *API) ensureRepositoryPaths(
	repo string, epoch string, database string, architecture string,
) error {
	var err error

	if _, err = os.Stat(repo); os.IsNotExist(err) {
		return fmt.Errorf("given repository doesn't exist")
	}

	if epoch == "" {
		return nil
	}

	if _, err = os.Stat(repo + "/" + epoch); os.IsNotExist(err) {
		return fmt.Errorf("given epoch doesn't exist")
	}

	if database == "" {
		return nil
	}

	if _, err = os.Stat(
		repo + "/" + epoch + "/" + database,
	); os.IsNotExist(err) {
		return fmt.Errorf("given database '%s' doesn't exist", database)
	}

	return nil
}

func (api *API) shouldChangePackageEpoch(context *gin.Context) bool {
	newEpoch := context.Request.Form.Get("new_epoch")
	if newEpoch == "" {
		return false
	}

	return true
}
