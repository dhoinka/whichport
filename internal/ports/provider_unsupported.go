//go:build !darwin && !linux

package ports

import (
	"context"
	"fmt"
	"runtime"
)

func discoverPlatform(_ context.Context, _ Query) ([]Listener, error) {
	return nil, fmt.Errorf("whichport does not support %s yet", runtime.GOOS)
}
