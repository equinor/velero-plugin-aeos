package plugin

import (
	"bufio"
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

func loadSecretsFile(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open %s", filepath)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		parseLineToEnv(scanner.Text())
	}
	return nil
}

func tryResolveSecretsFile(filepath string) (string, bool) {
	var err error
	var altFilename string

	if _, err = os.Stat(filepath); err == nil {
		return filepath, true
	}

	if _, exists := os.LookupEnv(secretsFileEnvVar); exists {
		altFilename = os.Getenv(secretsFileEnvVar)
		if altFilename != "" {
			return altFilename, true
		}
	}

	return "", false
}

func parseLineToEnv(text string) error {
	cleanedText := strings.TrimSpace(text)
	inputs := strings.SplitN(cleanedText, "=", 2)

	if isValidEnvVarName(inputs[0]) {
		os.Setenv(inputs[0], inputs[1])
	}
	return nil
}

func isValidEnvVarName(text string) bool {
	var output string = ""

	for _, x := range text {
		if unicode.IsUpper(x) || x == rune('_') {
			output = output + string(x)
		}
	}

	return text == output
}

func parseBlobDomainName(domainName string) string {
	return strings.Trim(
		strings.TrimSpace(
			strings.ToLower(domainName)), ".")
}

func validateEnvVars(envNames []string) bool {
	for _, name := range envNames {
		if _, ok := os.LookupEnv(name); !ok {
			return false
		}
	}
	return true
}
