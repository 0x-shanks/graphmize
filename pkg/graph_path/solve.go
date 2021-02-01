package graph_path

import (
	"path"
	"path/filepath"
)

func Solve(source string, currentDir string) (graphDir string) {
	if source != "" && !filepath.IsAbs(source) {
		graphDir = path.Join(currentDir, source)
	} else if filepath.IsAbs(source) {
		graphDir = source
	} else {
		graphDir = currentDir
	}
	return
}
