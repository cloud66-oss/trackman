#!/bin/bash

version=$(git branch | grep \* | cut -d ' ' -f2)
revision=$(git describe --tags --always)

echo "Building $version:$revision"
echo $version > VERSION
gox -ldflags "-X github.com/cloud66/trackman/utils.Revision=$revision -X github.com/cloud66/zonedns/utils.Version=$version" -os="darwin linux windows" -arch="amd64" -output "publish/{{.OS}}_{{.Arch}}_$version"
