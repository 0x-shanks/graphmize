package graph

import (
	"encoding/json"
	"fmt"
	"github.com/hourglasshoro/graphmize/pkg/file"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"os"
	"path"
	"strings"
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

func (g *Graph) ToTree() {
	var isLastLoopFlags []bool
	treeRecursion(g, isLastLoopFlags)
}

func treeRecursion(g *Graph, isLastLoopFlags []bool) {
	output(g.FileName, isLastLoopFlags)

	resources := g.Resources
	maxCount := len(resources)

	for i := 0; i < maxCount; i++ {
		isLastLoop := false
		if i == (maxCount - 1) {
			isLastLoop = true
		}
		flags := append(isLastLoopFlags, []bool{isLastLoop}...)
		treeRecursion(&resources[i], flags)
	}
}

func output(data string, isLastLoopFlags []bool) {
	pathLine := ""
	maxCount := len(isLastLoopFlags)
	for i := 0; i < maxCount; i++ {
		isLast := isLastLoopFlags[i]
		if i == (maxCount - 1) {
			if isLast {
				pathLine += "└── "
			} else {
				pathLine += "├── "
			}
		} else {
			if isLast {
				pathLine += "    "
			} else {
				pathLine += "│   "
			}
		}
	}
	pathLine += data
	fmt.Println(pathLine)
}

// Find determines if an element exists in the slice
func Find(slice []string, val string) (bool, int) {
	for i, item := range slice {
		if item == val {
			return true, i
		}
	}
	return false, -1
}

// BuildGraph recursively explores the specified directory, builds a dependency tree, and returns it
func BuildGraph(ctx file.Context, rootPath string) (*Graph, error) {
	rootGraph := NewGraph("root", "root", "/", []Graph{})

	// parentNodes determine if search should be skipped in BuildGraphFromDir
	parentNodes := map[string]*Graph{}

	// childNodes determine whether to put in parentNode
	childNodes := map[string]bool{}

	err := afero.Walk(ctx.FileSystem, rootPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			fileNameStartIndex := strings.LastIndex(path, "/")
			// If not rootPath
			if fileNameStartIndex > 0 {

				// Search kustomizationFile
				isKustomizationFile, _ := Find(file.KustomizationFileNames, path[fileNameStartIndex+1:])

				if isKustomizationFile {
					kustomizationFilePath := path[:fileNameStartIndex]
					kustomizationFile, err := ctx.GetKustomizationFromDirectory(kustomizationFilePath)
					if err != nil {
						return errors.Wrap(err, "cannot get kustomization file")
					}

					graph, err := BuildGraphFromDir(ctx, kustomizationFilePath, *kustomizationFile, parentNodes, childNodes)
					if err != nil {
						return errors.Wrap(err, "cannot get graph")
					}

					// Do not add already explored kustomization files to the parent
					if _, isChild := childNodes[kustomizationFilePath]; !isChild {
						parentNodes[kustomizationFilePath] = graph
					}
				}
			}

			return nil
		})
	if err != nil {
		return nil, err
	}

	for _, v := range parentNodes {
		rootGraph.Resources = append(rootGraph.Resources, *v)
	}

	return rootGraph, nil
}

// BuildGraphFromDir builds and returns a dependency tree from a kustomization file under the specified directory
func BuildGraphFromDir(ctx file.Context, directoryPath string, kustomizationFile file.KustomizationFile, parentNodes map[string]*Graph, childNodes map[string]bool) (*Graph, error) {
	var resources []Graph
	for _, resource := range kustomizationFile.Resources {

		resourcePath := path.Join(directoryPath, resource)
		isExist, err := afero.Exists(ctx.FileSystem, resourcePath)
		if err != nil {
			return nil, errors.Wrap(err, "cannot determine if resourcePath exist")
		}

		isDir, err := afero.IsDir(ctx.FileSystem, resourcePath)
		if !isExist || err != nil {
			resources = append(resources, *NewGraph("Unknown Resource", "Unknown Resource", resource, []Graph{}))
		} else if isDir {
			// For directories

			graph, isExplored := parentNodes[resourcePath]

			// If a file at this path is already registered as a parent when searching
			if isExplored {
				delete(parentNodes, resourcePath)
			} else {
				childKustomizationFile, err := ctx.GetKustomizationFromDirectory(resourcePath)
				if err != nil {
					return nil, errors.Wrap(err, "cannot get childKustomizationFile")
				}
				graph, err = BuildGraphFromDir(ctx, resourcePath, *childKustomizationFile, parentNodes, childNodes)
				if err != nil {
					return nil, errors.Wrap(err, "cannot buildGraph for childKustomizationFile")
				}
			}

			resources = append(resources, *graph)

			// Register nodes that have already been explored
			childNodes[resourcePath] = true

		} else if exist, _ := Find(file.KustomizationFileNames, resource); exist {
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
