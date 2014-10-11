// Package xattr provides a simple interface to user extended attributes on Linux and OSX.
// Support for xattrs is filesystem dependant, so not a given even if you are running one of those operating systems.
//
// On Linux you have to edit /etc/fstab to include "user_xattr". Also, Linux extended attributes have a manditory
// prefix of "user.". This is prepended transparently for Get/Set/Remove and hidden in List.
package xattr

// XAttrError records an error and the operation, file path and attribute that caused it.
type XAttrError struct {
	Op   string
	Path string
	Attr string
	Err  error
}

func (e *XAttrError) Error() string {
	return e.Op + " " + e.Path + " " + e.Attr + ": " + e.Err.Error()
}

// Returns whether the error is known to report that a extended attribute does not exist.
func IsNotExist(err error) bool {
	e, ok := err.(*XAttrError)
	if ok {
		return isNotExist(e)
	}
	return false
}

// Converts an array of NUL terminated UTF-8 strings
// to a []string.
func nullTermToStrings(buf []byte) (result []string) {
	offset := 0
	for index, b := range buf {
		if b == 0 {
			result = append(result, string(buf[offset:index]))
			offset = index + 1
		}
	}
	return
}

// Getxattr retrieves value of the extended attribute identified by attr
// associated with given path in filesystem into buffer dest.
//
// On success, dest contains data associated with attr, retrieved value size sz
// and nil error returned.
//
// On error, non-nil error returned. Getxattr returns error if dest was too
// small for attribute value.
//
// A nil slice can be passed as dest to get current size of attribute value,
// which can be used to estimate dest length for value associated with attr.
//
// Get is high-level function on top of Getxattr. Getxattr more efficient,
// because it issues one syscall per call, doesn't allocate memory for
// attribute data (caller can reuse buffer).
func Getxattr(path, attr string, dest []byte) (sz int, err error) {
	return get(path, attr, dest)
}

// Retrieves extended attribute data associated with path.
func Get(path, attr string) ([]byte, error) {
	attr = prefix + attr

	// find size
	size, err := Getxattr(path, attr, nil)
	if err != nil {
		return nil, &XAttrError{"getxattr", path, attr, err}
	}
	if size == 0 {
		return []byte{}, nil
	}

	// read into buffer of that size
	buf := make([]byte, size)
	size, err = Getxattr(path, attr, buf)
	if err != nil {
		return nil, &XAttrError{"getxattr", path, attr, err}
	}
	return buf[:size], nil
}

func Listxattr(path string, dest []byte) (sz int, err error) {
	return list(path, dest)
}

// Retrieves a list of names of extended attributes associated with path.
func List(path string) ([]string, error) {
	// find size
	size, err := Listxattr(path, nil)
	if err != nil {
		return nil, &XAttrError{"listxattr", path, "", err}
	}
	if size == 0 {
		return []string{}, nil
	}

	// read into buffer of that size
	buf := make([]byte, size)
	size, err = Listxattr(path, buf)
	if err != nil {
		return nil, &XAttrError{"listxattr", path, "", err}
	}
	return stripPrefix(nullTermToStrings(buf[:size])), nil
}

func Setxattr(path, attr string, data []byte, flags int) error {
	return set(path, attr, data, flags)
}

// Associates data as an extended attribute of path.
func Set(path, attr string, data []byte) error {
	attr = prefix + attr

	if err := Setxattr(path, attr, data, 0); err != nil {
		return &XAttrError{"setxattr", path, attr, err}
	}
	return nil
}

func Removexattr(path, attr string) error {
	return remove(path, attr)
}

// Removes the extended attribute.
func Remove(path, attr string) error {
	attr = prefix + attr
	if err := Removexattr(path, attr); err != nil {
		return &XAttrError{"removexattr", path, attr, err}
	}
	return nil
}
