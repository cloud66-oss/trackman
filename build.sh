#!/bin/bash

version=$(git describe --tags --always)
revision="dev"

echo "Building $version:$revision"
echo $version > publish/VERSION
gox -ldflags "-X github.com/cloud66/trackman/utils.Revision=$revision -X github.com/cloud66/zonedns/utils.Version=$version" -os="darwin linux windows" -arch="amd64" -output "publish/{{.OS}}_{{.Arch}}_$version"
