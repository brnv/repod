package main

import (
	"log"
	"net/http"
	"os"
	"path"
	"strings"

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
	Message string
	Data    string
	Status  int
}

func newAPI(root string) *API {
	return &API{
		root: root,
	}
}

func newAPIResponse() APIResponse {
	return APIResponse{
		Status: http.StatusOK,
	}
}

func (api *API) handleListRepositories(context *gin.Context) {
	response := newAPIResponse()

	repositories, err := listRepositories(api.root)
	if err != nil {
		response.Status = http.StatusInternalServerError
		response.Message = hierr.Errorf(
			err, "can't read repo dir %s", api.root,
		).Error()
	}

	response.Data = repositories

	api.sendResponse(context, response)
}

func (api *API) handleListPackages(context *gin.Context) {
	var (
		response = newAPIResponse()
		packages []string
		err      error
	)

	repository, err := getRepository(
		api.root,
		context.Query("path"),
		context.Query("system"),
	)
	if err != nil {
		response.Status = http.StatusBadRequest
		response.Message = hierr.Errorf(
			err,
			"can't start work with repo",
		).Error()
	}

	if repository != nil {
		packages, err = repository.ListPackages()
		if err != nil {
			response.Status = http.StatusInternalServerError
			response.Message = hierr.Errorf(
				err,
				"can't list packages for repo",
			).Error()
		}
	}

	response.Data = strings.Join(packages, "\n")

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

	repository, err := getRepository(
		api.root,
		context.Query("path"),
		context.Query("system"),
	)
	if err != nil {
		response.Status = http.StatusBadRequest
		response.Message = hierr.Errorf(
			err,
			"can't start work with repo",
		).Error()
	}

	if repository != nil {
		filename, file, err = api.copyFileFromRequest(repository, request)
		if err != nil {
			response.Message = hierr.Errorf(
				err,
				"can't get file from request",
			).Error()
		}
	}

	if file != nil {
		err = repository.AddPackage(filename, file, false)
		if err != nil {
			response.Message = hierr.Errorf(
				err,
				`can't add given package file %s`, filename,
			).Error()
		}
	}

	if len(response.Message) == 0 {
		response.Message = "package added"
	}

	api.sendResponse(context, response)
}

func (api *API) handleRemovePackage(context *gin.Context) {
	var (
		response    = newAPIResponse()
		packageName = context.Param("name")
		err         error
	)

	repository, err := getRepository(
		api.root,
		context.Query("path"),
		context.Query("system"),
	)
	if err != nil {
		response.Message = hierr.Errorf(
			err,
			"can't start work with repo",
		).Error()
	}

	if repository != nil {
		err = repository.RemovePackage(packageName)
		if err != nil {
			response.Message = err.Error()
		}
	}

	if len(response.Message) == 0 {
		response.Message = "package removed"
	}

	api.sendResponse(context, response)
}

func (api *API) handleEditPackage(context *gin.Context) {
	var (
		request     = context.Request
		packageName = context.Param("name")
		response    = newAPIResponse()
		file        *os.File
		filename    string
		err         error
	)

	request.ParseForm()
	rootNew := request.Form.Get("copy-to")

	repository, err := getRepository(
		api.root,
		context.Query("path"),
		context.Query("system"),
	)
	if err != nil {
		response.Message = hierr.Errorf(
			err,
			"can't start work with repo",
		).Error()
	}

	if rootNew != "" {
		filename, file, err = repository.GetPackageFile(packageName)
		repository.SetPath(rootNew)
	} else {
		filename, file, err = api.copyFileFromRequest(
			repository,
			context.Request,
		)
	}

	if err != nil {
		response.Message = hierr.Errorf(
			err,
			"can't prepare edit package",
		).Error()
	} else {
		err = repository.AddPackage(filename, file, true)
		if err != nil {
			response.Message = hierr.Errorf(
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
		packageName = context.Param("name")
	)

	repository, err := getRepository(
		api.root,
		context.Query("path"),
		context.Query("system"),
	)
	if err != nil {
		response.Message = hierr.Errorf(
			err,
			"can't start work with repo",
		).Error()
	}

	if repository != nil {
		response.Data, err = repository.DescribePackage(
			packageName,
		)
		if err != nil {
			response.Message = err.Error()
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
	if response.Message != "" {
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
