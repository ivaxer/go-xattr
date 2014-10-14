package xattr

import (
	"syscall"
)

func isNotExist(err *XAttrError) bool {
	return err.Err == syscall.ENOATTR
}
