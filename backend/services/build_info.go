package services

// BuildTime is the UTC time at which the binary was built. Release builds
// inject it with: -ldflags "-X sing_panel/services.BuildTime=<timestamp>".
var BuildTime = "unknown"
