package utils

import (
	"bufio"
	"os"
	"strings"
)

func ReadLines(filePath string, numberOfLines int) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	lines := make([]string, 0, numberOfLines)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		trimmedLine := strings.TrimSpace(scanner.Text())
		lines = append(lines, trimmedLine)
		if len(lines) >= numberOfLines {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func ReadLastLine(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}

	var line []byte
	var buf [1]byte
	pos := fileInfo.Size() - 1

	for pos >= 0 {
		_, err := file.ReadAt(buf[:], pos)
		if err != nil {
			return "", err
		}

		if buf[0] == '\n' && pos != fileInfo.Size()-1 {
			break
		}

		line = append([]byte{buf[0]}, line...)
		pos--
	}

	return strings.TrimRight(string(line), "\n"), nil
}
