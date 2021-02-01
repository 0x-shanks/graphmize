[![release](https://img.shields.io/github/v/release/hourglasshoro/graphmize?logo=github&style=for-the-badge)](https://github.com/hourglasshoro/graphmize/releases) [![codecov](https://img.shields.io/codecov/c/github/hourglasshoro/graphmize?logo=codecov&style=for-the-badge&token=CRSVAM7K1W)](https://codecov.io/gh/hourglasshoro/graphmize)

Graphmize is a tool to visualize the dependencies of kustomize.

# Installation
### Homebrew
```
brew tap hourglasshoro/homebrew-graphmize
brew install graphmize
```
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

# Example
The actual directory structure, manifest, etc. can be found on [this page](https://github.com/hourglasshoro/graphmize/tree/master/docs/example).
If you run Graphmize with this directory specified, the output will look like this.

```

example/overlays/production
└── example/base
    └── example/base/a_service
        ├── deployment.yaml
        │   └── example/overlays/production/a_service/deployment.yaml(p)
        └── service.yaml
            └── example/overlays/production/a_service/service.yaml(p)

example/overlays/staging
└── example/base
    └── example/base/a_service
        ├── deployment.yaml
        │   └── example/overlays/staging/a_service/deployment.yaml(p)
        └── service.yaml
```