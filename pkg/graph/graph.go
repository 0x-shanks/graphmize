package graph

import (
	"encoding/json"
	"github.com/hourglasshoro/graphmize/pkg/file"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"path"
)

type Graph struct {
	ApiVersion string  `json:"apiVersion"`
	Kind       string  `json:"kind"`
	FileName   string  `json:"fileName"`
	Resources  []Graph `json:"resources"`
	//Patches               []Graph
	//PatchesStrategicMerge []Graph
}

func NewGraph(
	apiVersion string,
	kind string,
	fileName string,
	resources []Graph,
) *Graph {
	graph := new(Graph)
	graph.ApiVersion = apiVersion
	graph.Kind = kind
	graph.FileName = fileName
	graph.Resources = resources
	return graph
}

// Marshal converts to json
func (g *Graph) Marshal() ([]byte, error) {
	result, err := json.Marshal(g)
	return result, err
}

// Find determines if an element exists in the slice
func Find(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// BuildGraph recursively calls resources from the root KustomizationFile to build a Graph
func BuildGraph(ctx file.Context, directoryPath string, kustomizationFile file.KustomizationFile) (*Graph, error) {
	var resources []Graph
	for _, resource := range kustomizationFile.Resources {

		resourcePath := path.Join(directoryPath, resource)
		isExist, err := afero.Exists(ctx.FileSystem, resourcePath)
		if err != nil {
			return nil, errors.Wrap(err, "cannot determine if resourcePath is a directory")
		}
		if !isExist {
			return nil, errors.New("the file or directory is not found")
		}

		isDir, err := afero.IsDir(ctx.FileSystem, resourcePath)
		if err != nil {
			return nil, errors.Wrap(err, "cannot determine if resourcePath is a directory")
		}
		if isDir {
			// For directories
			childKustomizationFile, err := ctx.GetKustomizationFromDirectory(resourcePath)
			if err != nil {
				return nil, errors.Wrap(err, "cannot get childKustomizationFile")
			}
			graph, err := BuildGraph(ctx, resourcePath, *childKustomizationFile)
			if err != nil {
				return nil, errors.Wrap(err, "cannot buildGraph for childKustomizationFile")
			}
			resources = append(resources, *graph)

		} else if exist := Find(file.KustomizationFileNames, resource); exist {
			// For kustomizationFile
			return nil, errors.New("must be a directory")
		} else {
			// If not kustomizationFile
			childResourceFile, err := ctx.GetResourceFromFile(resourcePath)
			if err != nil {
				return nil, errors.Wrap(err, "cannot get childResourceFile")
			}
			resources = append(resources, *NewGraph(childResourceFile.ApiVersion, childResourceFile.Kind, resource, []Graph{}))
		}
	}
	graph := NewGraph(kustomizationFile.ApiVersion, kustomizationFile.Kind, directoryPath, resources)
	return graph, nil
}
