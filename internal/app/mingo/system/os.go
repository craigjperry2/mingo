package system

import (
	"github.com/craigjperry2/mingo/internal/app/mingo/errors"
	"os"
	"os/user"
)

func Username() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", errors.ErrUnknownUser
	}
	return u.Username, nil
}

func Hostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", errors.ErrUnknownHost
	}
	return hostname, nil
}
