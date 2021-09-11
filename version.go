package xfsgo

import (
	"fmt"
	"runtime"
)

var version = "0.1.0"

func CurrentVersion() string {
	return version
}

func VersionString() string {
	program := "xfsgo"
	vs := "v" + CurrentVersion()
	osArch := runtime.GOOS + "/" + runtime.GOARCH
	return fmt.Sprintf("%s %s %s",
		program, vs, osArch)
}