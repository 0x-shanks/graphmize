package imput

import (
	"path"
	"path/filepath"
)

// Solve resolves the path of the root directory to be searched from the input source and currentDir
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
