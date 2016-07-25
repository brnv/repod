package main

type Repository interface {
	ListPackages() ([]string, error)
	AddPackage(repositoryPackage RepositoryPackage) error
	DeletePackage(repositoryPackage RepositoryPackage) error
	EditPackage(repositoryPackage RepositoryPackage) error
	DescribePackage(repositoryPackage RepositoryPackage) error
}

type RepositoryPackage string
