//go:build !windows

package lib

func readWindowsRegistryValue(_ string, _ string) (string, error) {
	return "", nil
}

func writeWindowsRegistryValue(_ string, _ string, _ string) error {
	return nil
}
