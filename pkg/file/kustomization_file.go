package file

import (
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
	"path"
)

// KustomizationFile represents a kustomization yaml file
type KustomizationFile struct {
	ApiVersion            string   `yaml:"apiVersion"`
	Kind                  string   `yaml:"kind"`
	Resources             []string `yaml:"resources"`
	PatchesStrategicMerge []string `yaml:"patchesStrategicMerge"`
}

// KustomizationFileNames represents a list of allowed filenames that
// kustomize searches for
var KustomizationFileNames = []string{
	"kustomization.yaml",
	"kustomization.yml",
	"Kustomization",
}

// NewFromFileSystem creates a context to interact with kustomization files from a provided file system
func NewFromFileSystem(fileSystem afero.Fs) *Context {
	return &Context{
		FileSystem: fileSystem,
	}
}

// GetKustomizationFromDirectory attempts to read a kustomization.yaml file from the given directory
func (c *Context) GetKustomizationFromDirectory(directoryPath string) (*KustomizationFile, error) {
	var kustomizationFile KustomizationFile

	fileUtility := &afero.Afero{Fs: c.FileSystem}

	fileFoundCount := 0
	kustomizationFilePath := ""
	for _, kustomizationFile := range KustomizationFileNames {
		currentPath := path.Join(directoryPath, kustomizationFile)

		exists, err := fileUtility.Exists(currentPath)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not check if file %v exists", currentPath)
		}

		if exists {
			kustomizationFilePath = currentPath
			fileFoundCount++
		}
	}

	if kustomizationFilePath == "" {
		return nil, errors.Wrapf(errors.New("Missing kustomization file"), "Error in directory %v", directoryPath)
	}

	if fileFoundCount > 1 {
		return nil, errors.Wrapf(errors.New("Too many kustomization files"), "Error in directory %v", directoryPath)
	}

	kustomizationFileBytes, err := fileUtility.ReadFile(kustomizationFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not read file %s", kustomizationFilePath)
	}

	err = yaml.Unmarshal(kustomizationFileBytes, &kustomizationFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not unmarshal yaml file %s", kustomizationFilePath)
	}

	return &kustomizationFile, nil
}
