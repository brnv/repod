package main

import (
	"io"
)

type Repository interface {
	ListPackages() ([]string, error)
	AddPackage(packageName string, file io.Reader) error
	RemovePackage(packageName string) error
	DescribePackage(packageName string) ([]string, error)
	ChangePackageEpoch(packageName string, epoch string) error
}
