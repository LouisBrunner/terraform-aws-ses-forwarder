//go:build tools

package build

import (
	_ "github.com/mattn/goveralls"
	_ "github.com/t-yuki/gocover-cobertura"
	_ "gotest.tools/gotestsum"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
