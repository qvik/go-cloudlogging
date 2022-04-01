# !/bin/sh

set -e

echo "Running unit tests.."
go test -v -bench=. github.com/qvik/go-cloudlogging
go test -v -bench=. github.com/qvik/go-cloudlogging/internal
