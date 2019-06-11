package builder

import (
	"io/ioutil"
	"strings"
)

func readFileToString(path string) (string, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func writeStringToFile(path string, content string) error {
	return ioutil.WriteFile(path, []byte(content), 0755)
}

func trimString(val string) string {
	return strings.Trim(val, " \r\n")
}
