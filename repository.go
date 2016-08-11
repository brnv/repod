package main

import (
	"io"
)

type Repository interface {
	ListPackages() ([]string, error)
	ListEpoches() ([]string, error)
	AddPackage(packageName string, file io.Reader, force bool) error
	RemovePackage(packageName string) error
	DescribePackage(packageName string) (string, error)
	EditPackage(packageName string, file io.Reader) error
	GetPackageFile(packageName string) (io.Reader, error)
	SetEpoch(epoch string)
}
