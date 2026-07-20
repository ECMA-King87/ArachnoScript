package lib

import (
	"os"
	"runtime"
)

// other platforms:
// const EOL = "\n"
//
// windows:
// EOL = "\r\n"
var EOL = "\n"

const DEBUG_MODE = true
const BUG_MESSAGE = "\x1b[31mUnexpected behaviour.\x1b[0m"
const PathListSepartor = os.PathListSeparator
const PathSepartor = os.PathSeparator

var NumCPU = runtime.NumCPU()

// darwin, (freebsd), linux, windows
const Platform = runtime.GOOS
const ARCH = runtime.GOARCH

func init() {
	if Platform == "windows" {
		EOL = "\r\n"
	}
}

func Version() string {
	return Sprintf("ArachnoScript Runtime Environment (ARE) \x1b[32m%s\x1b[0m%s%s on %s.%sParser version: %s%s", RUNTIME_VERSION, EOL, ToUpperCase(ARCH), Platform, EOL, PARSER_VERSION, EOL)
}

type Env struct {
	StrictMode bool
}

var ENV = Env{StrictMode: true}
