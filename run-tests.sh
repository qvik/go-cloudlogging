# !/bin/sh

set -e

echo "Running unit tests.."
go test -v github.com/qvik/go-cloudlogging
go test -v github.com/qvik/go-cloudlogging/internal
