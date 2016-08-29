#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

export CGO_ENABLED=0

DIRS="cmd pkg"
TARGETS=$(for d in ${DIRS}; do echo ./$d/...; done)

echo "Running tests:"
go test -i -installsuffix "static" ${TARGETS}
go test -installsuffix "static" ${TARGETS}
echo

echo -n "Checking gofmt: "
ERRS=$(find ${DIRS} -type f -name \*.go | xargs gofmt -l 2>&1 || true)
if [ -n "${ERRS}" ]; then
    echo "FAIL - the following files need to be gofmt'ed:"
    for e in ${ERRS}; do
        echo "    $e"
    done
    echo
    exit 1
fi
echo "PASS"
echo

echo -n "Checking go vet: "
ERRS=$(go vet ${TARGETS} 2>&1 || true)
if [ -n "${ERRS}" ]; then
    echo "FAIL"
    echo "${ERRS}"
    echo
    exit 1
fi
echo "PASS"
echo
