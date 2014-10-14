package xattr

import (
	"strings"
	"syscall"
)

func isNotExist(err *XAttrError) bool {
	return err.Err == syscall.ENODATA
}
