package plugin

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode"
)

func getRequiredSecrets(secretNames ...string) (map[string]string, error) {
	var secretsMap = make(map[string]string)

	for _, secretName := range secretNames {
		envVar, found := os.LookupEnv(secretName)
		if !found {
			return secretsMap, fmt.Errorf("required env var %s not set", secretName)
		}
		secretsMap[secretName] = envVar
	}

	return secretsMap, nil
}

func loadSecretsFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open %s", filename)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		parseLineToEnv(scanner.Text())
	}
	return nil
}

func parseLineToEnv(text string) error {
	cleanedText := strings.TrimSpace(text)
	inputs := strings.Split(cleanedText, "=")

	if len(inputs) != 2 {
		return errors.New("failed parse as valid env var assignment")
	}

	if isValidEnvVarName(inputs[0]) {
		os.Setenv(inputs[0], inputs[1])
	}
	return nil
}

func isValidEnvVarName(text string) bool {
	var output string = ""

	for _, x := range text {
		if unicode.IsUpper(x) {
			if unicode.IsLetter(x) || x == rune('_') {
				output = output + string(x)
			}
		}
	}

	if text == output {
		return true
	}
	return false
}

func parseBlobDomainName(domainName string) string {
	return strings.Trim(
		strings.TrimSpace(
			strings.ToLower(domainName)), ".")
}
