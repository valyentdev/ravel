package common

import (
	"errors"
	"fmt"
	"os/exec"
	"os/user"
	"strconv"
)

const RavelUser = "ravel-jailer"

// User represents the ravel-jailer user's UID and GID.
type User struct {
	Uid int
	Gid int
}

func lookupUser(name string) (*user.User, error) {
	u, err := user.Lookup(name)
	if err != nil {
		var e user.UnknownUserError
		if errors.As(err, &e) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to lookup user: %w", err)
	}
	return u, nil
}

func parseUidGid(u *user.User) (uid, gid int, err error) {
	uid, err = strconv.Atoi(u.Uid)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to convert uid to int: %w", err)
	}

	gid, err = strconv.Atoi(u.Gid)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to convert gid to int: %w", err)
	}

	return uid, gid, nil
}

// SetupRavelJailerUser ensures the ravel-jailer user exists and returns its UID/GID.
func SetupRavelJailerUser() (uid, gid int, err error) {
	u, err := lookupUser(RavelUser)
	if err != nil {
		return
	}
	if u != nil {
		return parseUidGid(u)
	}

	cmd := exec.Command("useradd", "-M", "-r", "-s", "/usr/sbin/nologin", RavelUser)
	err = cmd.Run()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create ravel-jailer user: %w", err)
	}

	u, err = lookupUser(RavelUser)
	if err != nil {
		return
	}

	return parseUidGid(u)
}
