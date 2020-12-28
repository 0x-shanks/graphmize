Graphmize is a tool to visualize the dependencies of kustomize.

# Installation
### Go
```
go get github.com/hourglasshoro/graphmize
```

# Usage
Run the graphmize command under the directory that contains the kustomization file.
```
graphmize
```

You can also specify a directory by using the source flag.However, only relative paths can be used.
```
graphmize -s [source path]
```