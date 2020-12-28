package graph

import (
	"errors"
	"github.com/hourglasshoro/graphmize/pkg/file"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"testing"
)

// BuildGraphFromDir Test

// TestBuildGraphRecursively tests to validate that when tree-structurally dependent
func TestBuildGraphRecursively(t *testing.T) {
	// Folder structure for this test
	//
	//   /app
	//   ├── kustomization.yaml
	//   └── sub
	//	   ├── kustomization.yaml
	//	   ├── a.yaml
	//	   └── b.yaml

	fake := afero.NewMemMapFs()
	ctx := file.NewContext(fake)
	fakeFileSystem := ctx.FileSystem
	fakeFileSystem.Mkdir("app", 0755)
	fakeFileSystem.Mkdir("app/sub", 0755)

	fileContents := `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- sub
`
	afero.WriteFile(fakeFileSystem, "app/kustomization.yaml", []byte(fileContents), 0644)

	fileContents = `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- a.yaml
- b.yaml
`
	afero.WriteFile(fakeFileSystem, "app/sub/kustomization.yaml", []byte(fileContents), 0644)

	fileContents = `
apiVersion: apps/v1
kind: Deployment
`
	afero.WriteFile(fakeFileSystem, "app/sub/a.yaml", []byte(fileContents), 0644)
	afero.WriteFile(fakeFileSystem, "app/sub/b.yaml", []byte(fileContents), 0644)

	dir := "app"
	kustomizationFile, _ := file.NewFromFileSystem(fakeFileSystem).GetKustomizationFromDirectory(dir)
	graph, err := BuildGraphFromDir(*ctx, "", dir, *kustomizationFile, map[string]*Graph{}, map[string]bool{})
	assert.Nil(t, err)

	expected := "a.yaml"
	actual := graph.Resources[0].Resources[0].FileName
	assert.Equal(t, expected, actual)

	expected = "b.yaml"
	actual = graph.Resources[0].Resources[1].FileName
	assert.Equal(t, expected, actual)

}

// TestBuildGraphRecursivelyJump tests to validate that specifying dependencies with relative paths
func TestBuildGraphRecursivelyJump(t *testing.T) {
	// Folder structure for this test
	//
	//   /app
	//   |
	//   ├── base
	//	 | ├── kustomization.yaml
	//	 | └── a.yaml
	//   |
	//   └── sub
	//	   ├── kustomization.yaml
	//	   └── b.yaml

	fake := afero.NewMemMapFs()
	ctx := file.NewContext(fake)
	fakeFileSystem := ctx.FileSystem
	fakeFileSystem.Mkdir("app", 0755)
	fakeFileSystem.Mkdir("app/base", 0755)
	fakeFileSystem.Mkdir("app/sub", 0755)

	fileContents := `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- a.yaml
`
	afero.WriteFile(fakeFileSystem, "app/base/kustomization.yaml", []byte(fileContents), 0644)

	fileContents = `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- ../base
- b.yaml
`
	afero.WriteFile(fakeFileSystem, "app/sub/kustomization.yaml", []byte(fileContents), 0644)

	fileContents = `
apiVersion: apps/v1
kind: Deployment
`
	afero.WriteFile(fakeFileSystem, "app/base/a.yaml", []byte(fileContents), 0644)
	afero.WriteFile(fakeFileSystem, "app/sub/b.yaml", []byte(fileContents), 0644)

	dir := "app/sub"
	kustomizationFile, _ := file.NewFromFileSystem(fakeFileSystem).GetKustomizationFromDirectory(dir)
	graph, err := BuildGraphFromDir(*ctx, "", dir, *kustomizationFile, map[string]*Graph{}, map[string]bool{})

	assert.Nil(t, err)

	expected := "a.yaml"
	actual := graph.Resources[0].Resources[0].FileName
	assert.Equal(t, expected, actual)

	expected = "b.yaml"
	actual = graph.Resources[1].FileName
	assert.Equal(t, expected, actual)

}

// BuildGraph Test

// TestBuildGraph tests to validate that root directory with no kustomization file
func TestBuildGraph(t *testing.T) {
	// Folder structure for this test
	//
	//   /app
	//   |
	//   ├── base
	//	 | ├── kustomization.yaml
	//	 | └── a.yaml
	//   |
	//   └── sub
	//	   ├── kustomization.yaml
	//	   └── b.yaml

	fake := afero.NewMemMapFs()
	ctx := file.NewContext(fake)
	fakeFileSystem := ctx.FileSystem
	fakeFileSystem.Mkdir("app", 0755)
	fakeFileSystem.Mkdir("app/base", 0755)
	fakeFileSystem.Mkdir("app/sub", 0755)

	fileContents := `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- a.yaml
`
	afero.WriteFile(fakeFileSystem, "app/base/kustomization.yaml", []byte(fileContents), 0644)

	fileContents = `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- ../base
- b.yaml
`
	afero.WriteFile(fakeFileSystem, "app/sub/kustomization.yaml", []byte(fileContents), 0644)

	fileContents = `
apiVersion: apps/v1
kind: Deployment
`
	afero.WriteFile(fakeFileSystem, "app/base/a.yaml", []byte(fileContents), 0644)
	afero.WriteFile(fakeFileSystem, "app/sub/b.yaml", []byte(fileContents), 0644)

	graph, err := BuildGraph(*ctx, "app")
	assert.Nil(t, err)

	expected := "a.yaml"
	actual := graph.Resources[0].Resources[0].Resources[0].FileName
	assert.Equal(t, expected, actual)

	expected = "b.yaml"
	actual = graph.Resources[0].Resources[1].FileName
	assert.Equal(t, expected, actual)

}

// TestBuildGraph tests to validate that when you have two independent dependencies
func TestBuildGraphWithTwoParents(t *testing.T) {
	// Folder structure for this test
	//
	//   /app
	//   |
	//   ├── base
	//	 | ├── kustomization.yaml
	//	 | └── a.yaml
	//   |
	//   └── staging
	//	 | ├── kustomization.yaml
	//	 | └── b.yaml
	//   └── production
	//	   ├── kustomization.yaml
	//	   └── c.yaml

	fake := afero.NewMemMapFs()
	ctx := file.NewContext(fake)
	fakeFileSystem := ctx.FileSystem
	fakeFileSystem.Mkdir("app", 0755)
	fakeFileSystem.Mkdir("app/base", 0755)
	fakeFileSystem.Mkdir("app/staging", 0755)
	fakeFileSystem.Mkdir("app/production", 0755)

	fileContents := `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- a.yaml
`
	afero.WriteFile(fakeFileSystem, "app/base/kustomization.yaml", []byte(fileContents), 0644)

	fileContents = `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- ../base
- b.yaml
`
	afero.WriteFile(fakeFileSystem, "app/staging/kustomization.yaml", []byte(fileContents), 0644)

	fileContents = `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- ../base
- c.yaml
`
	afero.WriteFile(fakeFileSystem, "app/production/kustomization.yaml", []byte(fileContents), 0644)

	fileContents = `
apiVersion: apps/v1
kind: Deployment
`
	afero.WriteFile(fakeFileSystem, "app/base/a.yaml", []byte(fileContents), 0644)
	afero.WriteFile(fakeFileSystem, "app/staging/b.yaml", []byte(fileContents), 0644)
	afero.WriteFile(fakeFileSystem, "app/production/c.yaml", []byte(fileContents), 0644)

	graph, err := BuildGraph(*ctx, "app")
	assert.Nil(t, err)

	expected := "a.yaml"
	actual := graph.Resources[0].Resources[0].Resources[0].FileName
	assert.Equal(t, expected, actual)

	for _, v := range graph.Resources {
		switch v.FileName {
		case "app/staging":
			expected = "b.yaml"
			actual = v.Resources[1].FileName
			assert.Equal(t, expected, actual)
		case "app/production":
			expected = "c.yaml"
			actual = v.Resources[1].FileName
			assert.Equal(t, expected, actual)
		default:
			assert.Error(t, errors.New("invalid File Name"))
		}
	}
}
