package lib

import (
	"os"
	"path/filepath"
)

func RealPath(path string) string {
	return catch(filepath.Abs(path))
}

func Rel(base, target string) string {
	return catch(filepath.Rel(base, target))
}

func IsAbs(path string) bool {
	return filepath.IsAbs(path)
}

func Abs(path string) string {
	return catch(filepath.Abs(path))
}

func JoinPaths(paths ...string) string {
	return filepath.Join(paths...)
}

// const exec_path = "C:\\Users\\ecmak\\ArachnoScript\\ARE\\v0.2\\ESK 2\\main.exe"

func ExecPath() string {
	// return exec_path
	return catch(filepath.EvalSymlinks(catch(os.Executable())))
}

func CWD() string {
	f := catch(os.Open("."))
	path := RealPath(catch(f.Stat()).Name())
	f.Close()
	return path
}

func DirOf(path string) string {
	return filepath.Dir(path)
}

func BaseName(path string) string {
	return filepath.Base(path)
}

func PathExt(path string) string {
	return filepath.Ext(path)
}

func CleanPath(path string) string {
	return filepath.Clean(path)
}

func EvalSymLink(path string) string {
	r, _ := filepath.EvalSymlinks(path)
	return r
}
