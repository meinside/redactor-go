# redactor-go

A go library for redacting private/sensitive information from given text, using OpenAI's [chat completion APIs](https://platform.openai.com/docs/api-reference/chat).

## Disclaimer

Chat completion APIs may miss some information, thus fail to redact such strings.

So results may vary even with the same input due to this limitation.

## Usage

With a client created with your OpenAI API key and organization ID,

```go
import "github.com/meinside/redactor-go"

const apiKey = "YOUR-OPENAI-API-KEY"
const orgID = "YOUR-OPENAI-ORGANIZATION-ID"
const model = "gpt-3.5-turbo"

if client, err := redactor.NewRedactor((&redactor.NewRedactorOptions{}).
        SetOpenAIAPIKeys(apiKey, orgID).
        SetModel(model)); err == nil {

    // TODO: do something here

}
```

### Detect private/sensitive information

```go
// detect strings and return them

const text = `some lengthy text with private/sensitive information`

if detected, err := client.Detect(text); err == nil {
    log.Printf("detected private/sensitive information: %+v", detected)
}
```

### Redact detected private/sensitive information

```go
// replace detected strings with given one
if redacted, err := client.Redact(text, "<REDACTED>"); err == nil {
    log.Printf("redacted text = %s", redacted)
}

// replace detected strings with the function results
if redacted, err := client.RedactFunc(text, func(in string) string {
    // NOTE: customize your redacted output here
    return fmt.Sprintf("<REDACTED:%d>", len(in))
}); err == nil {
    log.Printf("redacted text = %s", redacted)
}
```

## License

MIT

