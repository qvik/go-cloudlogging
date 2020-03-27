package cloudlogging

import (
	"testing"
)

func TestCreateAppEngineLogger(t *testing.T) {
	// Simply test compilation
	_ = MustNewAppEngineLogger()
}
