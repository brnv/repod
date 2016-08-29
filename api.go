package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/kovetskiy/toml"
	"github.com/reconquest/hierr"

	nucleus "git.rn/devops/nucleus-go"
)

type API struct {
	root     string
	authNeed bool
}

type APIResponse struct {
	Success bool
	Error   string
	Data    interface{}
	Status  int
}

func newAPI(root string) *API {
	return &API{
		root: root,
	}
}

func newAPIResponse() APIResponse {
	return APIResponse{
		Data:    make(map[string]interface{}),
		Status:  http.StatusOK,
		Success: true,
	}
}

func (api *API) handleListRepositories(context *gin.Context) {
	response := newAPIResponse()

	repositories, err := listRepositories(api.root)
	if err != nil {
		response.Status = http.StatusInternalServerError
		response.Error = hierr.Errorf(
			err, "can't read repo dir %s", api.root,
		).Error()
	}

	response.Data = repositories

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

	response.Data = epoches

	api.sendResponse(context, response)
}

func (api *API) handleListPackages(context *gin.Context) {
	var (
		response = newAPIResponse()
		packages []string
		err      error
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
		packages, err = repository.ListPackages()
		if err != nil {
			response.Status = http.StatusInternalServerError
			response.Error = hierr.Errorf(
				err,
				"can't list packages for repo",
			).Error()
		}
	}

	response.Data = packages

	api.sendResponse(context, response)
}

func (api *API) handleAddPackage(context *gin.Context) {
	var (
		request  = context.Request
		response = newAPIResponse()
		file     *os.File
		filename string
		err      error
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
		filename, file, err = api.copyFileFromRequest(repository, request)
		if err != nil {
			response.Error = hierr.Errorf(
				err,
				"can't get file from request",
			).Error()
		}
	}

	if file != nil {
		err = repository.AddPackage(filename, file, false)
		if err != nil {
			response.Error = hierr.Errorf(
				err,
				`can't add given package file %s`, filename,
			).Error()
		}
	}

	api.sendResponse(context, response)
}

func (api *API) handleRemovePackage(context *gin.Context) {
	var (
		response    = newAPIResponse()
		packageName = context.Param("package")
		err         error
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
		request     = context.Request
		packageName = context.Param("package")
		response    = newAPIResponse()
		file        *os.File
		filename    string
		err         error
	)

	request.ParseForm()
	pathNew := request.Form.Get("epoch_new")

	repository, err := api.newRepository(context)
	if err != nil {
		response.Error = hierr.Errorf(
			err,
			"can't start work with repo",
		).Error()
	}

	if pathNew != "" {
		filename, file, err = repository.GetPackageFile(packageName)
		repository.SetPath(pathNew)
	} else {
		filename, file, err = api.copyFileFromRequest(
			repository,
			context.Request,
		)
	}

	if err != nil {
		response.Error = hierr.Errorf(
			err,
			"can't prepare edit package",
		).Error()
	} else {
		err = repository.AddPackage(filename, file, true)
		if err != nil {
			response.Error = hierr.Errorf(
				err,
				"can't change package %s", packageName,
			).Error()
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
		response.Data, err = repository.DescribePackage(
			packageName,
		)
		if err != nil {
			response.Error = err.Error()
		}
	}

	api.sendResponse(context, response)
}

func (api *API) handleAuthentificate(context *gin.Context) {
	if !api.authNeed {
		return
	}

	_, token, ok := context.Request.BasicAuth()
	if !ok {
		context.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	user, err := nucleus.Authenticate(token)
	if err != nil {
		errorln(
			hierr.Errorf(
				err, "can't authentificate using token '%s'", token,
			),
		)
		context.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	context.Set("username", user.Name)
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
		repo         = context.Param("repo")
		epoch        = context.Param("epoch")
		database     = context.Param("db")
		architecture = context.Param("arch")
		repoPath     = filepath.Join(api.root, repo)
	)

	err := api.validateRepoPaths(repoPath, epoch, database, architecture)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			"can't validate repo paths: %s/%s/%s/%s",
			repoPath, epoch, database, architecture,
		)
	}

	return getRepository(
		api.root,
		filepath.Join(
			repo,
			epoch,
			database,
			architecture,
		),
	)
}

func (api *API) validateRepoPaths(
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

func (api *API) copyFileFromRequest(
	repository Repository,
	request *http.Request,
) (string, *os.File, error) {
	formFile, fileinfo, err := request.FormFile("package_file")
	if err != nil {
		return "", nil, hierr.Errorf(err, "can't read form file from request")
	}

	filePath, err := repository.CopyFileToRepo(
		path.Base(fileinfo.Filename),
		formFile,
	)
	if err != nil {
		return "", nil, hierr.Errorf(
			err,
			"can't copy file to repo",
		)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", nil, hierr.Errorf(
			err,
			"can't open file path %s", filePath,
		)
	}

	return filePath, file, nil
}
