package client

import (
	"os/user"
	"strings"
)

func ExpandFilePath(s string) string {
	if strings.HasPrefix(s, "~/") {
		usr, err := user.Current()
		if err != nil {
			return s
		}
		s = strings.Replace(s, "~", usr.HomeDir, 1)
	}
	return s
}
