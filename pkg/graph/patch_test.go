package graph

import (
	"github.com/hourglasshoro/graphmize/pkg/file"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestPatchesStrategicMerge is a test when PatchesStrategicMerge is specified in kustomize
func TestPatchesStrategicMerge(t *testing.T) {
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
	//	   └── patch.yaml

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

patchesStrategicMerge:
  - patch.yaml
`
	afero.WriteFile(fakeFileSystem, "app/sub/kustomization.yaml", []byte(fileContents), 0644)

	fileContents = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  replicas: 1
`

	afero.WriteFile(fakeFileSystem, "app/base/a.yaml", []byte(fileContents), 0644)

	fileContents = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  replicas: 3
`

	afero.WriteFile(fakeFileSystem, "app/sub/patch.yaml", []byte(fileContents), 0644)

	dir := "app/sub"
	kustomizationFile, _ := file.NewFromFileSystem(fakeFileSystem).GetKustomizationFromDirectory(dir)
	patchID := 0
	graph, err := BuildGraphFromDir(*ctx, "", dir, *kustomizationFile, &map[string]*Graph{}, &map[string]*Graph{}, &map[string]*Graph{}, &patchID)

	assert.Nil(t, err)

	expected := "app/sub/patch.yaml"
	actual := graph.Resources[0].Resources[0].Patches[0].FileName
	assert.Equal(t, expected, actual)
}
