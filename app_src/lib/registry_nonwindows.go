//go:build !windows

package lib

func readWindowsRegistryValue(keyPath string, valueName string) (string, error) {
	return "", nil
}

func writeWindowsRegistryValue(keyPath string, valueName string, value string) error {
	return nil
}
