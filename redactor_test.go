package redactor

import (
	"log"
	"os"
	"testing"
)

const (
	testModel = `gpt-3.5-turbo`

	testText = `This test script was written by a developer named John Doe. You can contact him through his email john.doe@no-such-domain.com or his cellphone number +1 123 456 7890. He is in South Pole, Antarctica right now, but will be back to answer you in a month or so.
`
)

func TestRedactor(t *testing.T) {
	_apiKey := os.Getenv("OPENAI_API_KEY")
	_org := os.Getenv("OPENAI_ORGANIZATION")
	_verbose := os.Getenv("VERBOSE")

	if len(_apiKey) <= 0 || len(_org) <= 0 {
		t.Errorf("environment variables `OPENAI_API_KEY` and `OPENAI_ORGANIZATION` are needed")
	}

	if client, err := NewRedactor((&NewRedactorOptions{}).
		SetOpenAIAPIKeys(_apiKey, _org).
		SetModel(testModel).
		SetVerbose(_verbose == "true")); err == nil {
		if redacted, err := client.Redact(testText, "<<<REDACTED>>>"); err != nil {
			t.Errorf("failed to redact text: %s", err)
		} else {
			if _verbose == "true" {
				log.Printf("redacted text = `%s`", redacted)
			}
		}
	} else {
		t.Errorf("failed to create a new redactor: %s", err)
	}
}

func TestRedactorFunc(t *testing.T) {
	_apiKey := os.Getenv("OPENAI_API_KEY")
	_org := os.Getenv("OPENAI_ORGANIZATION")
	_verbose := os.Getenv("VERBOSE")

	if len(_apiKey) <= 0 || len(_org) <= 0 {
		t.Errorf("environment variables `OPENAI_API_KEY` and `OPENAI_ORGANIZATION` are needed")
	}

	if client, err := NewRedactor((&NewRedactorOptions{}).
		SetOpenAIAPIKeys(_apiKey, _org).
		SetModel(testModel).
		SetVerbose(_verbose == "true")); err == nil {
		if redacted, err := client.RedactFunc(testText, func(from string) string {
			// NOTE: customize redacted strings here

			table := map[string]string{
				"John Doe":                    "<<<REDACTED-NAME>>>",
				"john.doe@no-such-domain.com": "<<<REDACTED-EMAIL>>>",
				"+1 123 456 7890":             "<<<REDACTED-MOBILE-NUMBER>>>",
			}
			if v, exists := table[from]; exists {
				return v
			}

			return "<<<REDACTED>>>"
		}); err != nil {
			t.Errorf("failed to redact text: %s", err)
		} else {
			if _verbose == "true" {
				log.Printf("redacted text = `%s`", redacted)
			}
		}
	} else {
		t.Errorf("failed to create a new redactor: %s", err)
	}
}
