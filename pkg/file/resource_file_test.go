package file

import (
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestGetResourceFromFile tests the GetResourceFromFile method to validate that the resource
// yaml file was marshaled correctly from the provided path
func TestGetResourceFromFile(t *testing.T) {
	// Folder structure for this test
	//
	//   /app
	//   └── deployment.yaml

	fakeFileSystem := afero.NewMemMapFs()
	fakeFileSystem.Mkdir("app", 0755)

	fileContents := `
apiVersion: apps/v1
kind: Deployment
`
	afero.WriteFile(fakeFileSystem, "app/deployment.yaml", []byte(fileContents), 0644)
	resourceFile, _ := NewFromFileSystem(fakeFileSystem).GetResourceFromFile("app/deployment.yaml")

	expected := "apps/v1"
	actual := resourceFile.ApiVersion
	assert.Equal(t, expected, actual)

	expected = "Deployment"
	actual = resourceFile.Kind
	assert.Equal(t, expected, actual)
}
