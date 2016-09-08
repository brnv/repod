package main

import (
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kovetskiy/toml"
	"github.com/reconquest/ser-go"

	nucleus "git.rn/devops/nucleus-go"
)

type API struct {
	root        string
	nucleusAuth bool
}

type APIResponse struct {
	Data  interface{}
	Error string

	successMessage string
}

func (api *API) detectRepository(context *gin.Context) {
	tracef("getting repository")

	var (
		path     = context.Query("path")
		system   = context.Query("system")
		response = context.MustGet("response").(*APIResponse)

		err error
	)

	if len(path) == 0 {
		return
	}

	if len(system) == 0 {
		system = getRepositorySystem(path)
	}

	debugf("path: %s, system: %s", path, system)

	repository, err := getRepository(api.root, path, system)
	if err != nil {
		context.AbortWithError(http.StatusInternalServerError, err)
		response.Error = ser.Errorf(err, "can't get repository").Error()
		api.sendResponse(context)
		return
	}

	if repository == nil {
		context.AbortWithError(http.StatusBadRequest, err)
		response.Error = ser.Errorf(err, "repository not detected").Error()
		api.sendResponse(context)
		return
	}

	debugf("repository: %s", repository)

	context.Set("repository", repository)
}

func (api *API) handleListRepositories(context *gin.Context) {
	tracef("listing repositories")

	response := context.MustGet("response").(*APIResponse)

	repositories, err := listRepositories(api.root)
	if err != nil {
		response.Error = ser.Errorf(err, "can't list repositories").Error()
	}

	response.Data = repositories

	api.sendResponse(context)
}

func (api *API) handleListPackages(context *gin.Context) {
	var (
		repository = context.MustGet("repository").(Repository)
		response   = context.MustGet("response").(*APIResponse)

		err error
	)

	tracef("listing packages")

	packages, err := repository.ListPackages()
	if err != nil {
		response.Error = ser.Errorf(err, "can't list packages").Error()
	}

	debugf("packages: %#v", packages)

	if len(packages) > 0 {
		response.Data = strings.Join(packages, "\n")
	}

	api.sendResponse(context)
}

func (api *API) handleAddPackage(context *gin.Context) {
	var (
		repository = context.MustGet("repository").(Repository)
		response   = context.MustGet("response").(*APIResponse)

		filename string
		err      error
	)

	tracef("saving user submitted file")

	filename, err = api.saveRequestFile(repository, context.Request)
	if err != nil {
		response.Error = ser.Errorf(
			err,
			"can't save file from request",
		).Error()
	}

	if len(filename) > 0 {
		tracef("adding package")

		err = repository.AddPackage(filename, forcePackageAdd)
		if err != nil {
			response.Error = ser.Errorf(
				err,
				`can't add package from`,
			).Error()
		}
	}

	response.successMessage = "package added"

	api.sendResponse(context)
}

func (api *API) handleRemovePackage(context *gin.Context) {
	var (
		repository  = context.MustGet("repository").(Repository)
		response    = context.MustGet("response").(*APIResponse)
		packageName = context.Param("name")

		err error
	)

	tracef("removing package")

	debugf("package name: %s", packageName)

	err = repository.RemovePackage(packageName)
	if err != nil {
		response.Error = ser.Errorf(err, "can't remove package").Error()
	}

	response.successMessage = "package removed"

	api.sendResponse(context)
}

func (api *API) handleEditPackage(context *gin.Context) {
	var (
		repository  = context.MustGet("repository").(Repository)
		response    = context.MustGet("response").(*APIResponse)
		packageName = context.Param("name")

		file     *os.File
		filename string
		err      error
	)

	debugf("package name: %s", packageName)

	tracef("parsing request form data")

	context.Request.ParseForm()
	pathCopyTo := context.Request.Form.Get("copy-to")

	if pathCopyTo == "" {
		tracef("saving user submitted file")

		filename, err = api.saveRequestFile(repository, context.Request)
		if err != nil {
			response.Error = ser.Errorf(
				err,
				"can't save file from request",
			).Error()
		}
	} else {
		tracef("getting package file directly from repository")

		file, err = repository.GetPackageFile(packageName)
		if err != nil {
			response.Error = ser.Errorf(
				err,
				"can't get package file",
			).Error()
		}

		if file != nil {
			filename = file.Name()
		}

		tracef("setting path to copy package to")

		repository.SetPath(pathCopyTo)
	}

	debugf("filename: %s", filename)

	if len(filename) > 0 {
		tracef("editing package")

		err = repository.AddPackage(filename, forcePackageEdit)
		if err != nil {
			response.Error = ser.Errorf(err, "can't edit package").Error()
		}
	}

	response.successMessage = "package edited"

	api.sendResponse(context)
}

func (api *API) handleDescribePackage(context *gin.Context) {
	var (
		repository  = context.MustGet("repository").(Repository)
		response    = context.MustGet("response").(*APIResponse)
		packageName = context.Param("name")
	)

	tracef("describing package")

	description, err := repository.DescribePackage(packageName)
	if err != nil {
		response.Error = ser.Errorf(err, "can't describe package").Error()
	}

	debugf("description: %s", description)

	response.Data = description

	api.sendResponse(context)
}

func (api *API) handleAuthentificate(context *gin.Context) {
	if !api.nucleusAuth {
		return
	}

	tracef("handle basic authorization")

	username, token, ok := context.Request.BasicAuth()
	if !ok {
		context.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	debugf("username: %s, token: %s", username, token)

	tracef("making nucleus authentication")

	user, err := nucleus.Authenticate(token)
	if err != nil {
		errorln(
			ser.Errorf(err, "can't authentificate"),
		)
		context.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	debugf("nucleus user: %#v", user)

	context.Set("username", user.Name)
}

func (api *API) sendResponse(context *gin.Context) {
	response := context.MustGet("response").(*APIResponse)

	debugf("response: %#v", response)

	if response.Data == nil && len(response.Error) == 0 {
		response.Data = response.successMessage
	}

	err := toml.NewEncoder(context.Writer).Encode(response)
	if err != nil {
		log.Printf("can't send response %#v", response)
		context.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *API) saveRequestFile(
	repository Repository, request *http.Request,
) (string, error) {
	tracef("getting form file")

	formfile, header, err := request.FormFile("package_file")
	if err != nil {
		return "", ser.Errorf(err, "can't read form file")
	}

	tracef("copying file to repository")

	pathCopied, err := repository.CopyFileToRepo(
		path.Base(header.Filename),
		formfile,
	)
	if err != nil {
		return "", ser.Errorf(err, "can't copy file to repo")
	}

	debugf("saved file: %s", pathCopied)

	return pathCopied, nil
}

func (api *API) prepareResponse(context *gin.Context) {
	context.Set("response", &APIResponse{})
}
