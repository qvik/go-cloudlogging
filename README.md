# Cloud Logging for Go

[![GoDoc](https://godoc.org/github.com/qvik/go-cloudlogging?status.svg)](https://godoc.org/github.com/qvik/go-cloudlogging)

This module provides a log wrapper that is intended to handle logging in cloud-based backend environment.

## Changelog

- 1.1.0: Renamed Stackdriver -> Google Cloud Logging; this is API-breaking change.
- 1.0.0: Bumped version to 1.0.0. Fixed zap logger accumulating structured logging fields.
- 0.0.16: Added JSON formatting output hints
- 0.0.15: Added default parameters for structured logging
- 0.0.13: Argument handling bugfix, added GoDoc reference to README
- 0.0.11: Improved documentation.

## Usage

Install the dependency:

```sh
go get -u github.com/qvik/go-cloudlogging
```

## Convenience constructors & examples

For convenience, several constructor methods are provided; see below.

_Google Compute Engine (GCE) example:_

```go
func init() {
	log = cloudlog.MustNewComputeEngineLogger("project-id", "MyAppLog")
}
```

This could also used for Kuhernetes.

_Google App Engine (GAE) ecample:_

```go
func init() {
	log = cloudlog.MustNewAppEngineLogger() // Optionally define log ID as arg
}
```

_Google Cloud Functions (GCF) example:_

```go
func init() {
	log = cloudlog.MustNewCloudFunctionLogger() // Optionally define log ID as arg
}
```

_AWS Elastic Beanstalk / EC2 example:_

```go
	logfile := "/var/log/example-app.log"
	log = cloudlogging.MustNewLogger(cloudlogging.WithZap(),
		cloudlogging.WithOutputHints(cloudlogging.JSONFormat),
		cloudlogging.WithOutputPaths(logfile),
		cloudlogging.WithErrorOutputPaths(logfile))
}
```

## License

The library is distributed with the [MIT License](LICENSE.md).
