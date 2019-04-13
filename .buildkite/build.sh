#!/bin/bash

version=$(git describe --tags --always)
force="false"

if [[ $FORCE == "--force" ]]
  then
    force="true"
fi

if [[ $BUILDKITE_BRANCH -eq "master" ]]
  then
    channel="edge"
  else
    channel="stable"
fi

mkdir build

docker run -i --rm -w /gopath/src/github.com/cloud66/trackman -v $(pwd):/gopath/src/github.com/cloud66/trackman cloud66/gobuildchain /bin/bash << COMMANDS
gox -ldflags "-X github.com/cloud66/trackman/utils.Version=$version -X github.com/cloud66/trackman/utils.Channel=$channel" -os="darwin linux windows" -arch="amd64" -output "build/{{.OS}}_{{.Arch}}_$version"
chown -R 999:998 build
COMMANDS
