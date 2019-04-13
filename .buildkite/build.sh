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

docker run -w /gopath/src/github.com/cloud66/trackman -v /var/lib/buildkite-agent/builds/buildkite-cloud66-com-1/cloud-66/trackman:/gopath/src/github.com/cloud66/trackman cloud66/gobuildchain gox -ldflags "-X github.com/cloud66/trackman/utils.Version=$version -X github.com/cloud66/trackman/utils.Channel=$channel" -os="darwin linux windows" -arch="amd64" -output "build/{{.OS}}_{{.Arch}}_$version"
