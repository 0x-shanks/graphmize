package graph

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/hourglasshoro/graphmize/pkg/file"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Graph represents a node that is a customization file or resource file
type Graph struct {
	ApiVersion string   `json:"apiVersion"`
	Kind       string   `json:"kind"`
	FileName   string   `json:"fileName"`
	Resources  []*Graph `json:"resources"`
	Patches    map[int]*Graph
}

// NewGraph is Graph constructor
func NewGraph(
	apiVersion string,
	kind string,
	fileName string,
	resources []*Graph,
	patches map[int]*Graph,
) *Graph {
	graph := new(Graph)
	graph.ApiVersion = apiVersion
	graph.Kind = kind
	graph.FileName = fileName
	graph.Resources = resources
	graph.Patches = patches
	return graph
}

// Marshal converts to json
func (g *Graph) Marshal() ([]byte, error) {
	result, err := json.Marshal(g)
	return result, err
}

// ToTree displays a tree structure
func (g *Graph) ToTree() {
	treeRecursion(g, []bool{}, g.Patches, true)
}

// treeRecursion calls output for each hierarchy
func treeRecursion(g *Graph, isLastLoopFlags []bool, patches map[int]*Graph, isRoot bool) {
	output(g.FileName, isLastLoopFlags, false)

	for i, patch := range g.Patches {
		_, ok := patches[i]
		if ok && !isRoot {
			output(patch.FileName+"(p)", append(isLastLoopFlags, []bool{true}...), true)
		}
	}

	resources := g.Resources
	maxCount := len(resources)

	for i := 0; i < maxCount; i++ {
		isLastLoop := false
		if i == (maxCount - 1) {
			isLastLoop = true
		}
		flags := append(isLastLoopFlags, []bool{isLastLoop}...)
		treeRecursion(resources[i], flags, patches, false)
	}
}

// output prints the result to the standard output
func output(data string, isLastLoopFlags []bool, isPatch bool) {
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
	if isPatch {
		c := color.New(color.FgCyan)
		fmt.Print(pathLine)
		_, _ = c.Println(data)
	} else {
		pathLine += data
		fmt.Println(pathLine)
	}
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

	rootGraph := NewGraph("root", "root", "/", []*Graph{}, nil)

	// parentNodes determine if search should be skipped in BuildGraphFromDir; map[resourcePath]*Node
	parentNodes := map[string]*Graph{}

	// childNodes determine whether to put in parentNode; map[resourcePath]*Node
	childNodes := map[string]*Graph{}

	// resourceNode is the data to determine the patch; map[metadata.name]*Node
	resourceNodes := map[string]*Graph{}
	// patchID is an Id to identify the patch that appeared
	patchID := 0

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

					graph, err := BuildGraphFromDir(ctx, rootPath, kustomizationFilePath, *kustomizationFile, &parentNodes, &childNodes, &resourceNodes, &patchID)
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
		rootGraph.Resources = append(rootGraph.Resources, v)
	}

	return rootGraph, nil
}

// BuildGraphFromDir builds and returns a dependency tree from a kustomization file under the specified directory
func BuildGraphFromDir(ctx file.Context, rootPath string, directoryPath string, kustomizationFile file.KustomizationFile, parentNodesPtr *map[string]*Graph, childNodesPtr *map[string]*Graph, resourceNodesPtr *map[string]*Graph, patchID *int) (*Graph, error) {
	var resources []*Graph

	parentNodes := *parentNodesPtr
	childNodes := *childNodesPtr
	resourceNodes := *resourceNodesPtr

	for _, resource := range kustomizationFile.Resources {

		resourcePath := path.Join(directoryPath, resource)
		isExist, err := afero.Exists(ctx.FileSystem, resourcePath)
		if err != nil {
			return nil, errors.Wrap(err, "cannot determine if resourcePath exist")
		}

		isDir, err := afero.IsDir(ctx.FileSystem, resourcePath)
		if !isExist || err != nil {
			resources = append(resources, NewGraph("Unknown Resource", "Unknown Resource", resource, []*Graph{}, nil))
		} else if isDir {
			// For directories

			// If a file at this path is already registered as a parent when searching
			if graph, isParent := parentNodes[resourcePath]; isParent {
				delete(parentNodes, resourcePath)
				resources = append(resources, graph)

				// Register nodes that have already been explored
				childNodes[resourcePath] = graph
			} else if graph, isChild := childNodes[resourcePath]; isChild {
				// If a file at this path is already registered as a child when searching
				resources = append(resources, graph)
			} else {
				childKustomizationFile, err := ctx.GetKustomizationFromDirectory(resourcePath)
				if err != nil {
					return nil, errors.Wrap(err, "cannot get childKustomizationFile")
				}
				graph, err := BuildGraphFromDir(ctx, rootPath, resourcePath, *childKustomizationFile, parentNodesPtr, childNodesPtr, resourceNodesPtr, patchID)
				if err != nil {
					return nil, errors.Wrap(err, "cannot buildGraph for childKustomizationFile")
				}
				resources = append(resources, graph)

				// Register nodes that have already been explored
				childNodes[resourcePath] = graph
			}

		} else if exist, _ := Find(file.KustomizationFileNames, resource); exist {
			// For kustomizationFile
			return nil, errors.New("must be a directory")
		} else {
			// If not kustomizationFile
			childResourceFile, err := ctx.GetResourceFromFile(resourcePath)
			if err != nil {
				return nil, errors.Wrap(err, "cannot get childResourceFile")
			}
			graph := NewGraph(childResourceFile.ApiVersion, childResourceFile.Kind, resource, []*Graph{}, map[int]*Graph{})
			resources = append(resources, graph)
			// If the patch has already been found when searching for the kustomization file
			resource, exist := resourceNodes[childResourceFile.Metadata.Name]
			if exist {
				// Patch IDs is stored in resource.Patch
				graph.Patches = resource.Patches
			}
			resourceNodes[childResourceFile.Metadata.Name] = graph
		}
	}

	// Store the patchID; map[patchID]*Node
	patches := map[int]*Graph{}

	// Explore the paths passed by PatchesStrategicMerge
	for _, patch := range kustomizationFile.PatchesStrategicMerge {
		patchPath := path.Join(directoryPath, patch)
		_, err := afero.Exists(ctx.FileSystem, patchPath)
		if err != nil {
			return nil, errors.Wrap(err, "cannot determine if patchPath exist")
		}
		patchResourceFile, err := ctx.GetResourceFromFile(patchPath)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get patchResourceFile")
		}
		resource, ok := resourceNodes[patchResourceFile.Metadata.Name]

		formRootPath, err := filepath.Rel(rootPath, patchPath)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get patch path from root")
		}

		patchGraph := NewGraph(patchResourceFile.ApiVersion, patchResourceFile.Kind, formRootPath, []*Graph{}, map[int]*Graph{})

		if ok {
			// When the resource has already been registered
			resource.Patches[*patchID] = patchGraph
		} else {
			// When a resource is not registered
			resourceNodes[patchResourceFile.Metadata.Name] = NewGraph("", "", "", []*Graph{}, patches)
		}

		patches[*patchID] = patchGraph
		*patchID++
	}

	relPath, err := filepath.Rel(rootPath, directoryPath)
	if err != nil {
		return nil, err
	}
	graph := NewGraph(kustomizationFile.ApiVersion, kustomizationFile.Kind, relPath, resources, patches)
	return graph, nil
}
