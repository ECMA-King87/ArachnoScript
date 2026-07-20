//go:build windows

package lib

import "golang.org/x/sys/windows/registry"

func readWindowsRegistryValue(keyPath string, valueName string) (string, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer key.Close()

	value, _, err := key.GetStringValue(valueName)
	if err != nil {
		return "", err
	}
	return value, nil
}

func writeWindowsRegistryValue(keyPath string, valueName string, value string) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	return key.SetStringValue(valueName, value)
}
