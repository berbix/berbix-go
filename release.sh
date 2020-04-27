#!/bin/bash

set -ex

VERSION=$(cat version)

sed -i "" -e "s/sdkVersion = \"[[:digit:]]*\.[[:digit:]]*\.[[:digit:]]*\"/sdkVersion = \"$VERSION\"/g" berbix.go

git add *.go go.mod version
git commit -m "Updating Berbix Go SDK version to $VERSION"
git tag -a $VERSION -m "Version $VERSION"
git push --follow-tags
