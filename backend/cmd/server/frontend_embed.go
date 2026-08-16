package main

import "embed"

// frontendFiles embeds the built frontend into the server binary.
// The release/CI build copies frontend/dist to cmd/server/web/dist before
// compiling the backend, so the binary remains self-contained.
//
//go:embed web/dist
var frontendFiles embed.FS
