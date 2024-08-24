package dotenv

import (
	"errors"
	"os"
	"reflect"
	"testing"
)

var (
	validCommentsMap = map[string]string{
		"foo": "bar",
		"baz": "foo",
		"bar": "foo",
	}
	validQuotedMap = map[string]string{
		"OPTION_A": "1",
		"OPTION_B": "2",
		"OPTION_C": "",
		"OPTION_D": "\\n",
		"OPTION_E": "1",
		"OPTION_F": "2",
		"OPTION_G": "",
		"OPTION_H": "\n",
		"OPTION_I": "echo 'asd'",
		"OPTION_J": "line 1\nline 2",
		"OPTION_K": "    line one\nthis is 'quoted'\none more line",
		"OPTION_L": "line 1\nline 2",
		"OPTION_M": "line one\nthis is \"quoted\"\none more line",
	}
	validExportPrefixMap = map[string]string{
		"OPTION_A": "2",
		"OPTION_B": "exported",
	}

	validSubstitutionMap = map[string]string{
		"OPTION_A": "1",
		"OPTION_B": "1",
		"OPTION_C": "1",
		"OPTION_D": "11",
		"OPTION_E": "",
	}

	validMultiLineMap = map[string]string{
		"KEY": "Hi, my name is\nEyad.",
	}
)

func TestRead(t *testing.T) {
	readEnvTests := []struct {
		name     string
		filename string
		want     map[string]string
		err      error
	}{
		{
			name:     "invalid extension file",
			filename: "../fixtures/invalid/invalid_extension.json",
			want:     nil,
			err:      inValidFileExtension,
		},
		{
			name:     "valid file extension",
			filename: "../fixtures/valid/comments.env",
			want:     validCommentsMap,
		},
	}

	for _, tt := range readEnvTests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Read(tt.filename)
			if !errors.Is(err, tt.err) {
				t.Fatalf("expected error %v but got %v", tt.err, err)
			}

			assertMaps(t, got, tt.want)
		})
	}
}

func TestParse(t *testing.T) {
	parseInvalidEnvTests := []struct {
		name     string
		filename string
		want     map[string]string
		err      error
	}{
		{
			name:     "invalid line",
			filename: "../fixtures/invalid/invalid.env",
			err:      inValidLine,
		},
		{
			name:     "invalid key",
			filename: "../fixtures/invalid/invalid_key.env",
			err:      inValidKey,
		},
		{
			name:     "unterminated multiline value",
			filename: "../fixtures/invalid/unterminated_multiline.env",
			err:      unterminatedMultiLine,
		},
		{
			name:     "unterminated quoted value",
			filename: "../fixtures/invalid/unterminated_quote.env",
			err:      unterminatedQuote,
		},
		{
			name:     "unexpected characters after value",
			filename: "../fixtures/invalid/unexpected_characters.env",
			err:      unexpectedCharacters,
		},
	}

	parseValidEnvTests := []struct {
		name     string
		filename string
		want     map[string]string
	}{
		{
			name:     "valid comments",
			filename: "../fixtures/valid/comments.env",
			want:     validCommentsMap,
		},
		{
			name:     "valid quoted values",
			filename: "../fixtures/valid/quoted.env",
			want:     validQuotedMap,
		},
		{
			name:     "valid multiline value",
			filename: "../fixtures/valid/multiline.env",
			want:     validMultiLineMap,
		},
		{
			name:     "using export before key",
			filename: "../fixtures/valid/exported.env",
			want:     validExportPrefixMap,
		},
		{
			name:     "substituting variables",
			filename: "../fixtures/valid/substitutions.env",
			want:     validSubstitutionMap,
		},
	}

	for _, tt := range parseInvalidEnvTests {
		t.Run(tt.name, func(t *testing.T) {
			envFile, err := os.Open(tt.filename)
			assertNoError(t, err)
			defer envFile.Close()
			_, err = Parse(envFile)
			assertError(t, err, tt.err)
		})
	}

	for _, tt := range parseValidEnvTests {
		t.Run(tt.name, func(t *testing.T) {
			envFile, err := os.Open(tt.filename)
			assertNoError(t, err)
			defer envFile.Close()
			got, err := Parse(envFile)
			assertNoError(t, err)
			assertMaps(t, got, tt.want)
		})
	}
}

func TestLoad(t *testing.T) {
	loadEnvTests := []struct {
		name     string
		filename string
		want     map[string]string
		err      error
	}{
		{
			name:     "invalid extension file",
			filename: "../fixtures/invalid/invalid_extension.json",
			want:     nil,
			err:      inValidFileExtension,
		},
		{
			name:     "valid file extension",
			filename: "../fixtures/valid/comments.env",
			want:     validCommentsMap,
		},
	}

	for _, tt := range loadEnvTests {
		t.Run(tt.name, func(t *testing.T) {
			err := Load(tt.filename)
			assertError(t, err, tt.err)

			for key, val := range tt.want {
				if os.Getenv(key) != val {
					t.Fatalf("expected %s=%s but got %s=%s", key, val, key, os.Getenv(key))
				}
			}

			t.Cleanup(func() {
				os.Clearenv()
			})
		})
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected nil but got %v", err)
	}
}

func assertError(t *testing.T, got error, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Fatalf("expected error %v but got %v", want, got)
	}
}

func assertMaps(t *testing.T, got, want map[string]string) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v but got %v", want, got)
	}
}
