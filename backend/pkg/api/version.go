package api

import "time"

// Version is set at build time via -ldflags "-X ragpack/pkg/api.Version=x.y.z".
// Falls back to "dev" for local builds.
var Version = "dev"

var startedAt = time.Now()
