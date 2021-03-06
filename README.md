# Cloud Logging for Go

[![GoDoc](https://godoc.org/github.com/qvik/go-cloudlogging?status.svg)](https://godoc.org/github.com/qvik/go-cloudlogging)

This module provides a log wrapper that is intended to handle logging in cloud-based backend environment. It can be configured to use a local logger or a over-the-network cloud logging library. Currently the only supported cloud logger is Stackdriver. For local logging purposes, the very efficient Zap logging library is used. It can also be used in AWS environments to produce queryable JSON formatted logs to be inspected in CloudWatch.

The library is intended to be used with Go 1.10+ with module support but does not require that.

The module provides convenience constructor methods for various possible deployments Google Cloud Platform. Similar support for AWS is planned and pull requests for such are appreciated.

## Changelog

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

Import the library and define a logger variable:

```go
import cloudlog "github.com/qvik/go-cloudlogging"

var (
	log *cloudlog.Logger
)
```

Then, typically in your module's `init()` method, create your logger.

Use the `NewLogger()` function to create a new logger. It provides a lot of flexibility in configuring the logger you need.

For example, to create a logger that logs to both local and Stackdriver logger, you could do something like:

```go
func init() {
	opts := []cloudlog.LogOption{}
	opts = append(opts, cloudlog.WithStackdriver(projectID, "", logID, nil))
	opts = append(opts, cloudlog.WithZap())

	log, err := cloudlog.NewLogger(opts)
	if err != nil {
		panic(fmt.Sprintf("failed to create a logger: %v", err))
	}
}
```

For convenience, several constructor methods are provided; see below.

### Google Cloud Platform convenience methods

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
func init() {
	logfile := "/var/log/example-app.log"
	log = cloudlogging.MustNewLogger(cloudlogging.WithZap(),
		cloudlogging.WithOutputHints(cloudlogging.JSONFormat),
		cloudlogging.WithOutputPaths(logfile),
		cloudlogging.WithErrorOutputPaths(logfile))
}
```

## License

The library is distributed with the MIT License.

## Contributing

Contributions to this library are welcomed. Any contributions have to meet the following criteria:

- Meaningfulness. Discuss whether what you are about to contribute indeed belongs to this library in the first place before submitting a pull request.
- Code style. Use gofmt and golint and you cannot go wrong with this. Generally do not exceed a line length of 80 characters.
- Testing. Try and include tests for your code.

## Contact

Any questions? Contact matti@qvik.fi.

## References

- [Stackdriver Logging](https://cloud.google.com/logging/)
- [Stackdriver Go API](https://godoc.org/cloud.google.com/go/logging)
- [Zap Logger](https://github.com/uber-go/zap)
