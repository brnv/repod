package main

import (
	"errors"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kovetskiy/toml"
	"github.com/reconquest/ser-go"

	"git.rn/devops/nucleus-go"
)

type API struct {
	root        string
	authEnabled bool
}

type APIResponse struct {
	Data  interface{}
	Error string

	message string
}

func (api *API) readRequestParams(context *gin.Context) {
	debugf("getting repository")

	var (
		path     = context.Query("path")
		system   = context.Query("system")
		force    = context.Query("force")
		response = context.MustGet("response").(*APIResponse)

		err error
	)

	if path == "" {
		return
	}

	if system == "" {
		system = getRepositorySystem(path)
	}

	context.Set("force", false)
	if force == "1" {
		context.Set("force", true)
	}

	tracef("path: %s, system: %s", path, system)

	repository, err := getRepository(api.root, path, system)
	if err != nil {
		response.Error = ser.Errorf(err, "can't get repository").Error()
		context.AbortWithError(http.StatusInternalServerError, err)
		sendResponse(response, context)
		return
	}

	if repository == nil {
		response.Error = ser.Errorf(
			err,
			"repository is not detected",
		).Error()
		context.AbortWithError(http.StatusNotFound, err)
		sendResponse(response, context)
		return
	}

	tracef("repository: %s", repository)

	context.Set("repository", repository)

}

func (api *API) handleListRepositories(context *gin.Context) {
	debugf("listing repositories")

	response := context.MustGet("response").(*APIResponse)

	repositories, err := listRepositories(api.root)
	if err != nil {
		response.Error = ser.Errorf(err, "can't list repositories").Error()
	}

	response.Data = repositories

	sendResponse(response, context)
}

func (api *API) handleListPackages(context *gin.Context) {
	var (
		repository, exists = context.Get("repository")
		response           = context.MustGet("response").(*APIResponse)
	)

	if !exists {
		response.Error = errors.New("repository is not defined").Error()
	} else {
		debugf("listing packages")

		packages, err := repository.(Repository).ListPackages()
		if err != nil {
			response.Error = ser.Errorf(err, "can't list packages").Error()
		}

		tracef("packages: %#v", packages)

		if len(packages) > 0 {
			response.Data = strings.Join(packages, "\n")
		}
	}

	sendResponse(response, context)
}

func (api *API) handleAddPackage(context *gin.Context) {
	var (
		repository, exists = context.Get("repository")
		response           = context.MustGet("response").(*APIResponse)

		filename string
		err      error
	)

	if !exists {
		response.Error = errors.New("repository is not defined").Error()
	} else {
		debugf("saving user submitted file")

		filename, err = api.saveRequestFile(
			repository.(Repository),
			context.Request,
		)
		if err != nil {
			response.Error = ser.Errorf(
				err,
				"can't save file from request",
			).Error()
		}
	}

	if filename != "" {
		debugf("adding package")

		err = repository.(Repository).AddPackage(
			filename,
			context.MustGet("force").(bool),
		)
		if err != nil {
			response.Error = ser.Errorf(err, `can't add package`).Error()
		}
	}

	response.message = "package added"

	sendResponse(response, context)
}

func (api *API) handleRemovePackage(context *gin.Context) {
	var (
		repository = context.MustGet("repository").(Repository)
		response   = context.MustGet("response").(*APIResponse)

		name = context.Param("name")

		version = context.Query("version")

		err error
	)

	debugf("removing package")

	tracef("package name: %s", name)

	err = repository.RemovePackage(name, version)
	if err != nil {
		response.Error = ser.Errorf(err, "can't remove package file").Error()
	}

	response.message = "package removed"

	sendResponse(response, context)
}

func (api *API) handleCopyPackage(context *gin.Context) {
	var (
		repository  = context.MustGet("repository").(Repository)
		response    = context.MustGet("response").(*APIResponse)
		packageName = context.Param("name")

		packageVersion = context.Query("package_version")
		pathNew        = context.Query("copy-to")
	)

	err := repository.CopyPackage(packageName, packageVersion, pathNew)
	if err != nil {
		response.Error = ser.Errorf(err, "can't edit package").Error()
	}

	response.message = "package copied"

	sendResponse(response, context)
}

func (api *API) handleDescribePackage(context *gin.Context) {
	var (
		repository  = context.MustGet("repository").(Repository)
		response    = context.MustGet("response").(*APIResponse)
		packageName = context.Param("name")
	)

	debugf("describing package")

	description, err := repository.DescribePackage(packageName)
	if err != nil {
		response.Error = ser.Errorf(err, "can't describe package").Error()
	}

	tracef("description: %s", description)

	response.Data = description

	sendResponse(response, context)
}

func (api *API) handleAuthentificate(context *gin.Context) {
	if !api.authEnabled {
		return
	}

	debugf("handle basic authorization")

	username, token, ok := context.Request.BasicAuth()
	if !ok {
		context.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	debugf("nucleus authentication")

	user, err := nucleus.Authenticate(token)
	if err != nil {
		errorln(
			ser.Errorf(err, "can't authentificate with nucleus"),
		)
		context.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if user.Name != username {
		errorln(
			ser.Errorf(
				err,
				"nucleus user not matched with request user",
			),
		)
		context.AbortWithStatus(http.StatusUnauthorized)
		return
	}
}

func (api *API) saveRequestFile(
	repository Repository, request *http.Request,
) (string, error) {
	debugf("getting form file")

	formfile, header, err := request.FormFile("package_file")
	if err != nil {
		return "", ser.Errorf(err, "can't read form file")
	}

	debugf("copying file to repository")

	pathCopied, err := repository.CreatePackageFile(
		path.Base(header.Filename),
		formfile,
	)
	if err != nil {
		return "", ser.Errorf(err, "can't copy file to repo")
	}

	tracef("saved file: %s", pathCopied)

	return pathCopied, nil
}

func (api *API) prepareResponse(context *gin.Context) {
	context.Set("response", &APIResponse{})
}

func sendResponse(response *APIResponse, context *gin.Context) {
	if response.Data == nil && len(response.Error) == 0 {
		response.Data = response.message
	}

	tracef("response: %#v", response)

	err := toml.NewEncoder(context.Writer).Encode(response)
	if err != nil {
		errorln(err)
		context.AbortWithError(http.StatusInternalServerError, err)
	}

	if len(response.Error) > 0 {
		errorln(response.Error)
	}
}
