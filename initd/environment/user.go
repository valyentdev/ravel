package environment

import (
	"os/user"
	"strconv"
)

func resolveUser(username string) (uid int, gid int, homeDir string, err error) {
	u, err := user.Lookup(username)
	if err != nil {
		return
	}

	uid, err = strconv.Atoi(u.Uid)
	if err != nil {
		return
	}

	gid, err = strconv.Atoi(u.Gid)
	if err != nil {
		return
	}

	return uid, gid, u.HomeDir, nil
}
