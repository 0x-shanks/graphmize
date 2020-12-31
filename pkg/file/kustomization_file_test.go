package file

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/spf13/afero"
)

// TestNoKustomizationFiles tests to validate that when no kustomization files
// are found, an error is returned
func TestNoKustomizationFiles(t *testing.T) {
	// Folder structure for this test
	//
	//   /app

	fakeFileSystem := afero.NewMemMapFs()
	fakeFileSystem.Mkdir("app", 0755)
	_, err := NewFromFileSystem(fakeFileSystem).GetKustomizationFromDirectory("app")

	assert.NotNil(t, err)
}

// TestMultipleKustomizationFiles tests to validate that when multiple kustomization files
// are found, an error is returned
func TestMultipleKustomizationFiles(t *testing.T) {
	// Folder structure for this test
	//
	//   /app
	//   ├── kustomization.yaml
	//   └── kustomization.yml

	fakeFileSystem := afero.NewMemMapFs()
	fakeFileSystem.Mkdir("app", 0755)
	emptyFileContents := ""

	afero.WriteFile(fakeFileSystem, "app/kustomization.yaml", []byte(emptyFileContents), 0644)
	afero.WriteFile(fakeFileSystem, "app/kustomization.yml", []byte(emptyFileContents), 0644)
	_, err := NewFromFileSystem(fakeFileSystem).GetKustomizationFromDirectory("app")

	assert.NotNil(t, err)
}

// TestGetFromDirectory tests the GetFromDirectory method to validate that the kustomization
// yaml file was marshaled correctly from the provided path
func TestGetFromDirectory(t *testing.T) {
	// Folder structure for this test
	//
	//   /app
	//   └── kustomization.yaml

	fakeFileSystem := afero.NewMemMapFs()
	fakeFileSystem.Mkdir("app", 0755)

	fileContents := `
resources:
- a.yaml
`
	afero.WriteFile(fakeFileSystem, "app/kustomization.yaml", []byte(fileContents), 0644)
	kustomizationFile, _ := NewFromFileSystem(fakeFileSystem).GetKustomizationFromDirectory("app")

	expected := "a.yaml"
	actual := kustomizationFile.Resources[0]

	assert.Equal(t, expected, actual)
}
