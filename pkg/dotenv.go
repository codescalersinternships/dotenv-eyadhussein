package dotenv

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

var (
	inValidFileExtension  = errors.New("invalid file extension")
	inValidLine           = errors.New("invalid line")
	inValidKey            = errors.New("invalid key")
	unterminatedMultiLine = errors.New("unterminated multiline value")
	unterminatedQuote     = errors.New("unterminated quoted value")
	unexpectedCharacters  = errors.New("unexpected characters after value")
)

var (
	keyRegex           = regexp.MustCompile(`^[a-zA-Z_]+[a-zA-Z0-9_]*`)
	substituteVarRegex = regexp.MustCompile(`(\\)?(\$)(\()?\{?([A-Z0-9_]+)?\}?`)
	escapeRegex        = regexp.MustCompile(`\\.`)
	unescapeRegex      = regexp.MustCompile(`\\([^$])`)
)

// Read reads the environment variables from the given files and returns them as a map.
func Read(filenames ...string) (map[string]string, error) {
	envVars := make(map[string]string)

	for _, filename := range filenames {
		if path.Ext(filename) != ".env" {
			return nil, fmt.Errorf("%w for file %s", inValidFileExtension, filename)
		}

		envFile, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer envFile.Close()

		log.Println("reading file", filename)
		vars, err := Parse(envFile)
		if err != nil {
			return nil, err
		}

		for key, val := range vars {
			envVars[key] = val
		}
	}

	return envVars, nil
}

// Parse parses the environment variables from the given file as a reader and returns them as a map.
func Parse(envFile io.Reader) (map[string]string, error) {
	envVars := make(map[string]string)

	scanner := bufio.NewScanner(envFile)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#") || len(strings.TrimSpace(line)) == 0 {
			continue
		}

		line = strings.TrimPrefix(line, "export ")
		keyVal := strings.SplitN(line, "=", 2)

		if len(keyVal) != 2 {
			return nil, inValidLine
		}

		key := strings.TrimSpace(keyVal[0])
		if matched := keyRegex.MatchString(key); !matched {
			return nil, fmt.Errorf("%w for %s", inValidKey, key)
		}
		val, err := extractValue(keyVal[1], scanner, envVars)
		if err != nil {
			return nil, fmt.Errorf("failed to parse value for key %s %w", key, err)
		}

		envVars[key] = val
	}
	return envVars, nil
}

// Load loads the environment variables from the given files into the current environment.
func Load(filenames ...string) error {
	envVars, err := Read(filenames...)
	if err != nil {
		return err
	}

	for key, val := range envVars {
		if err := os.Setenv(key, val); err != nil {
			return err
		}
	}

	return nil
}

func extractValue(val string, scanner *bufio.Scanner, currentEnvVars map[string]string) (string, error) {
	if !strings.HasPrefix(val, "'") && !strings.HasPrefix(val, "\"") {
		return parseEscape(substituteVariables(strings.TrimSpace(strings.Split(val, "#")[0]), currentEnvVars)), nil
	}

	var remaining string

	if isMultiLine(val) {
		var multilineVal strings.Builder

		var prefix string
		if strings.HasPrefix(val, "\"\"\"") {
			prefix = "\"\"\""
		} else {
			prefix = "'''"
		}
		for scanner.Scan() {
			line := scanner.Text()

			if strings.HasSuffix(line, prefix) {
				return strings.TrimSuffix(multilineVal.String(), "\n"), nil
			}

			line = strings.TrimSuffix(line, prefix)
			line = strings.TrimSuffix(line, prefix)

			if prefix == "\"\"\"" {
				line = substituteVariables(line, currentEnvVars)
			}
			multilineVal.WriteString(parseEscape(line) + "\n")
		}

		return "", unterminatedMultiLine
	} else {
		var prefix byte

		if strings.HasPrefix(val, "'") {
			prefix = '\''
		} else {
			prefix = '"'
		}
		if strings.HasPrefix(val, string(prefix)) {
			for i := 1; i < len(val); i++ {
				if val[i] == prefix && !(i > 0 && val[i-1] == '\\') {
					val, remaining = val[1:i], strings.TrimSpace(val[i+1:])
				}
			}

			if strings.HasPrefix(val, string(prefix)) {
				return "", unterminatedQuote
			}
		}

		if prefix == '"' {
			val = parseEscape(substituteVariables(val, currentEnvVars))
		}
	}

	if strings.TrimSpace(remaining) != "" && !strings.HasPrefix(remaining, "#") {
		return "", unexpectedCharacters
	}

	return val, nil
}

func isMultiLine(line string) bool {
	return strings.HasPrefix(line, "\"\"\"") || strings.HasPrefix(line, "'''")
}

func parseEscape(str string) string {
	out := escapeRegex.ReplaceAllStringFunc(str, func(match string) string {
		c := strings.TrimPrefix(match, `\`)
		switch c {
		case "n":
			return "\n"
		case "r":
			return "\r"
		case "t":
			return "\t"
		case "f":
			return "\f"
		case "b":
			return "\b"
		default:
			return match
		}
	})
	return unescapeRegex.ReplaceAllString(out, "$1")
}

func substituteVariables(line string, envVars map[string]string) string {
	return substituteVarRegex.ReplaceAllStringFunc(line, func(match string) string {
		if _, ok := envVars[match[2:len(match)-1]]; !ok {
			return ""
		}
		return envVars[match[2:len(match)-1]]
	})
}
