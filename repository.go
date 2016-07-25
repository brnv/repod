package main

import (
	"io"
)

type Repository interface {
	ListPackages() ([]string, error)
	AddPackage(filename string, file io.Reader) error
	DeletePackage(repositoryPackage RepositoryPackage) error
	EditPackage(repositoryPackage RepositoryPackage) error
	DescribePackage(repositoryPackage RepositoryPackage) error
}

type RepositoryPackage struct {
	Name string
	File io.Reader
}
