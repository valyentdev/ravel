package vm

import (
	"errors"
	"fmt"
	"os/exec"
	"os/user"
	"strconv"
)

const RAVEL_USER = "ravel-jailer"

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

func setupRavelJailerUser() (uid, gid int, err error) {
	user, err := lookupUser(RAVEL_USER)
	if err != nil {
		return
	}
	if user != nil {
		return parseUidGid(user)
	}

	cmd := exec.Command("useradd", "-M", "-r", "-s", "/usr/sbin/nologin", RAVEL_USER)
	err = cmd.Run()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create ravel-jailer user: %w", err)
	}

	user, err = lookupUser(RAVEL_USER)
	if err != nil {
		return
	}

	return parseUidGid(user)
}
