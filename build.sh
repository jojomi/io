#!/bin/bash
set -ex

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd "${DIR}" || exit 1

pushd "cmd/io" > /dev/null

GIT_COMMIT=$(git rev-list -1 HEAD)
GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
GIT_VERSION=$(git tag --list | grep "^v" | tail -n 1)
GIT_DATE=$(git show -s --format=%ci HEAD)
GIT_STATE=$(git diff --quiet && echo 'clean' || echo 'dirty')
GIT_REMOTE=$(git config --get remote.origin.url)

# trying to be reproducible
go install -trimpath -ldflags "-buildid= -X 'main.GitCommit=$GIT_COMMIT' -X 'main.GitBranch=$GIT_BRANCH' -X 'main.GitDate=$GIT_DATE' -X 'main.GitVersion=$GIT_VERSION' -X 'main.GitState=$GIT_STATE'  -X 'main.GitRemote=$GIT_REMOTE'" .

popd > /dev/null

# build README for Github
mmdc -i docu/overview.mmd -o docu/overview.svg
io --allow-exec --template docu/README.tpl.md --output README.md