//go:build tools
// +build tools

package tools

import (
	_ "github.com/bokwoon95/wgo"                            // Live reload
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint" // Linter
	_ "golang.org/x/tools/cmd/goimports"                    // goimports
	_ "honnef.co/go/tools/cmd/staticcheck"                  // Staticcheck
	_ "mvdan.cc/gofumpt"                                    // gofumpt
)
