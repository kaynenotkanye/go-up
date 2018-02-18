package reader

import (
	"os"
	"bufio"
	"strings"
)

type FileReader struct {
	filename string
	ignoreNotFound bool
}

func (f *FileReader) Read() (map[string]Property, error) {

	config := map[string]Property{}

	file, err := os.Open(f.filename)
	if err != nil {
		if f.ignoreNotFound {
			return config, nil
		} else {
			return nil, err
		}
	}
	defer file.Close()


	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if equal := strings.Index(line, "="); equal >= 0 {
			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
				value := ""
				if len(line) > equal {
					value = strings.TrimSpace(line[equal+1:])
				}
				config[key] = Property{value, true}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return config, nil
}