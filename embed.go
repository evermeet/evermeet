package main

import "embed"

//go:embed all:web/build
var webFS embed.FS
