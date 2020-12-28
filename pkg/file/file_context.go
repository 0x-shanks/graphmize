package file

import "github.com/spf13/afero"

type Context struct {
	FileSystem afero.Fs
}

// NewContext returns a new context to interact with files
func NewContext(fileSystem afero.Fs) *Context {
	return NewFromFileSystem(fileSystem)
}
