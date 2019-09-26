# Cloud Logging for Go

This module provides a log wrapper that handles the most commong GCP deployment envinments and provides a standard compatible logging API.

Usage example:

```go
import cloudlog "github.com/qvik/go-cloudlogging"

var (
	log *cloudlog.Logger
)

func init() {
	log = cloudlog.MustNewAppEngineLogger().SetLogLevel(cloudlog.Debug)
}
```