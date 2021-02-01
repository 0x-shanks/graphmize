package file

import (
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

// ResourceFile represents any files except kustomization yaml file
type ResourceFile struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
}

// GetResourceFromFile attempts to read a yaml file from the given file name
func (c *Context) GetResourceFromFile(resourcePath string) (*ResourceFile, error) {
	fileUtility := &afero.Afero{Fs: c.FileSystem}
	fileBytes, err := fileUtility.ReadFile(resourcePath)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not read file %s", resourcePath)
	}

	var resourceFile ResourceFile
	err = yaml.Unmarshal(fileBytes, &resourceFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not unmarshal yaml file %s", resourcePath)
	}
	return &resourceFile, nil
}
