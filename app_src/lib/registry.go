package lib

import (
	"bufio"
	"os"
	"strings"
)

func ReadRegistryValue(keyPath string, valueName string) (string, error) {
	if Platform == "windows" {
		return readWindowsRegistryValue(keyPath, valueName)
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
		return writeWindowsRegistryValue(keyPath, valueName, value)
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
