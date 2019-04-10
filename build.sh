#!/bin/bash

version=$(git describe --tags --always)

if [ -z "$1" ]
  then
    echo "No channel supplied"
    exit 1
fi

channel=$1

echo "Building $channel/$version"
echo

rm build/*
curl -s http://downloads.cloud66.com.s3.amazonaws.com/trackman/versions.json | jq '.versions |= map(if (.channel == "'$channel'") then .version = "'$version'" else . end)' > build/versions.json
echo "Current Versions"
cat build/versions.json | jq -r '.versions | map([.channel, .version] | join(": ")) | .[]'
echo

gox -ldflags "-X github.com/cloud66/trackman/utils.Version=$version -X github.com/cloud66/trackman/utils.Channel=$channel" -os="darwin linux windows" -arch="amd64" -output "build/{{.OS}}_{{.Arch}}_$version"
