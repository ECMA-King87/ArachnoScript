package lib

import (
	"bufio"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func ReadRegistryValue(keyPath string, valueName string) (string, error) {
	if Platform == "windows" {
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
	home, err := UserHomeDir()
	if err != nil {
		return "", err
	}
	targetFile := JoinPaths(home, ".bashrc")
	file, err := os.Open(targetFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "export") && strings.Contains(line, valueName) {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				cleanValue := strings.Trim(parts[1], "\"")
				return cleanValue, nil
			}
		}
	}

	return "", nil
}

func WriteRegistryValue(keyPath string, valueName string, value string) error {
	if Platform == "windows" {
		key, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.SET_VALUE)
		if err != nil {
			return err
		}
		defer key.Close()

		return key.SetStringValue(valueName, value)
	}

	home, err := UserHomeDir()
	if err != nil {
		return err
	}
	targetFile := JoinPaths(home, ".bashrc")
	file, err := OpenFileWithFlags(targetFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY)
	if err != nil {
		targetFile = JoinPaths(home, ".bash_profile")
		file, err = OpenFileWithFlags(targetFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY)
		if err != nil {
			targetFile = JoinPaths(home, ".zshrc")
			file, err = OpenFileWithFlags(targetFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY)
			if err != nil {
				return err
			}
		}
	}
	defer file.Close()
	envLine := "export " + valueName + "=" + keyPath + "\n"
	_, err = file.WriteString(envLine)
	if err != nil {
		return err
	}
	return nil
}
