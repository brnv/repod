package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kovetskiy/toml"
	"github.com/reconquest/hierr"
)

type API struct {
	repositoriesDir string
	repositoryOS    string
}

type APIResponse struct {
	Success bool
	Error   string
	Data    map[string][]string
	Status  int
}

const (
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
		Status:  http.StatusOK,
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

func (api *API) handleListRepositories(context *gin.Context) {
	response := newAPIResponse()

	repositories, err := ioutil.ReadDir(api.repositoriesDir)
	if err != nil {
		response.Status = http.StatusInternalServerError
		response.Error = hierr.Errorf(
			err, "can't read repo dir %s", api.repositoriesDir,
		).Error()
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
		response = newAPIResponse()
		epoches  []string
	)

	repository, err := api.newRepository(context)
	if err != nil {
		response.Status = http.StatusBadRequest
		response.Error = hierr.Errorf(
			err,
			"can't start work with repo",
		).Error()
	}

	if repository != nil {
		epoches, err = repository.ListEpoches()
		if err != nil {
			response.Status = http.StatusInternalServerError
			response.Error = hierr.Errorf(
				err,
				"can't list epoches for repo",
			).Error()
		}
	}

	if len(epoches) == 0 && response.Error == "" {
		response.Status = http.StatusBadRequest
		response.Error = fmt.Errorf("no epoches found for repo").Error()
	}

	for _, epoch := range epoches {
		response.Data[sliceKeyEpoches] = append(
			response.Data[sliceKeyEpoches], epoch,
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
		response.Status = http.StatusBadRequest
		response.Error = hierr.Errorf(
			err,
			"can't start work with repo",
		).Error()
	}

	if repository != nil {
		response.Data[sliceKeyPackages], err = repository.ListPackages()
		if err != nil {
			response.Status = http.StatusInternalServerError
			response.Error = hierr.Errorf(
				err,
				"can't list packages for repo",
			).Error()
		}
	}

	api.sendResponse(context, response)
}

func (api *API) handleAddPackage(context *gin.Context) {
	var (
		err         error
		response    = newAPIResponse()
		request     = context.Request
		packageName = context.Param("package")
		file        io.Reader
	)

	repository, err := api.newRepository(context)
	if err != nil {
		response.Status = http.StatusBadRequest
		response.Error = hierr.Errorf(
			err,
			"can't start work with repo",
		).Error()
	}

	if repository != nil {
		file, err = api.getFileFromRequest(request)
		if err != nil {
			response.Error = hierr.Errorf(
				err,
				"can't get file from request",
			).Error()
		}
	}

	if file != nil {
		err = repository.AddPackage(packageName, file, false)
		if err != nil {
			response.Error = hierr.Errorf(
				err,
				"can't add package",
			).Error()
		}
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
		response.Error = hierr.Errorf(
			err,
			"can't start work with repo",
		).Error()
	}

	if repository != nil {
		err = repository.RemovePackage(packageName)
		if err != nil {
			response.Error = err.Error()
		}
	}

	api.sendResponse(context, response)
}

func (api *API) handleEditPackage(context *gin.Context) {
	var (
		err         error
		response    = newAPIResponse()
		request     = context.Request
		packageName = context.Param("package")
		file        io.Reader
	)

	request.ParseForm()
	epochNew := request.Form.Get("epoch_new")

	repository, err := api.newRepository(context)
	if err != nil {
		response.Error = hierr.Errorf(
			err,
			"can't start work with repo",
		).Error()
	}

	if repository != nil {
		if epochNew != "" {
			file, err = repository.GetPackageFile(packageName)
			repository.SetEpoch(epochNew)
		} else {
			file, err = api.getFileFromRequest(request)
		}

		if err != nil {
			response.Error = err.Error()
		}

		err = repository.EditPackage(packageName, file)
		if err != nil {
			response.Error = err.Error()
		}
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
		response.Error = hierr.Errorf(
			err,
			"can't start work with repo",
		).Error()
	}

	if repository != nil {
		response.Data[sliceKeyPackage], err = repository.DescribePackage(
			packageName,
		)
		if err != nil {
			response.Error = err.Error()
		}
	}

	api.sendResponse(context, response)
}

func (api *API) sendResponse(context *gin.Context, response APIResponse) {
	if response.Error != "" {
		response.Success = false
		if response.Status == 0 {
			response.Status = http.StatusInternalServerError
		}
	}

	err := toml.NewEncoder(context.Writer).Encode(response)
	if err != nil {
		log.Printf("can't send response %#v", response)
		context.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	context.Writer.WriteHeader(response.Status)
}

func (api *API) newRepository(context *gin.Context) (Repository, error) {
	var (
		repoPath = filepath.Join(
			api.repositoriesDir,
			context.Param("repo"),
		)
		epoch        = context.Param("epoch")
		database     = context.Param("db")
		architecture = context.Param("arch")
	)

	err := api.ensureRepositoryPaths(repoPath, epoch, database, architecture)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			"can't ensure repo paths: %s/%s/%s/%s",
			repoPath, epoch, database, architecture,
		)
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

	return nil, fmt.Errorf("can't parse repository from request")
}

func (api *API) ensureRepositoryPaths(
	repo string, epoch string, database string, architecture string,
) error {
	if _, err := os.Stat(repo); err != nil {
		return hierr.Errorf(err, "can't stat repo %s", repo)
	}

	if epoch == "" {
		return nil
	}

	if _, err := os.Stat(filepath.Join(repo, epoch)); err != nil {
		return hierr.Errorf(
			err, "can't stat repo's epoch %s/%s", repo, epoch,
		)
	}

	if database == "" {
		return nil
	}

	if _, err := os.Stat(filepath.Join(repo, epoch, database)); err != nil {
		return hierr.Errorf(
			err, "can't stat repo's epoch's database %s/%s/%s",
			repo, epoch, database,
		)
	}

	return nil
}

func (api *API) getFileFromRequest(request *http.Request) (io.Reader, error) {
	file, _, err := request.FormFile(postFormPackageFile)
	if err != nil {
		return nil, hierr.Errorf(err, "can't read form file from request")
	}

	return file, nil
}
