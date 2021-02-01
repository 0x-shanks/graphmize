package imput

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Solve Test

//  TestSolveWithNoSource tests to validate when not specified in the -s option
func TestSolveWithNoSource(t *testing.T) {
	source := ""
	actual := Solve(source, "/User/test//app")
	expected := "/User/test//app"
	assert.Equal(t, expected, actual)
}

//  TestSolveWithNoSource tests to validate when a relative path is specified with the -s option
func TestSolveWithRelativePath(t *testing.T) {
	source := "../app2"
	actual := Solve(source, "/User/test/app1")
	expected := "/User/test/app2"
	assert.Equal(t, expected, actual)
}

//  TestSolveWithNoSource tests to validate when a absolute path is specified with the -s option
func TestSolveWithAbsolutePath(t *testing.T) {
	source := "/User/test/app2"
	actual := Solve(source, "/app1")
	expected := "/User/test/app2"
	assert.Equal(t, expected, actual)
}
