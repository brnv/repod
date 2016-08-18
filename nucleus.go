package main

import (
	nucleus "git.rn/devops/nucleus-go"
	"github.com/kovetskiy/hierr"
)

type NucleusAuth struct {
	Address string `default:"_nucleus"`
}

func (auth *NucleusAuth) AddCertificateFile(tlsCert string) error {
	err := nucleus.AddCertificateFile(tlsCert)
	if err != nil {
		return hierr.Errorf(
			err,
			"can't add nucleus server certificate",
		)
	}

	nucleus.SetAddress(auth.Address)
	nucleus.SetUserAgent("missiond/" + version)

	return nil
}
