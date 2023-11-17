package redactor

import (
	"fmt"
	"strings"

	"github.com/meinside/openai-go"
)

const (
	defaultModel = "gpt-3.5-turbo"

	chatCompletionTemperature = 0.1 // FIXME: need to be changed for better accuracy

	userAgent = "redactor-go"
)

// Redactor client struct
type Redactor struct {
	openAI *openAIConfigs
}

// OpenAI API key and organization
type openAIConfigs struct {
	apiKey       *string
	organization *string
	model        *string

	client *openai.Client

	verbose bool
}

type NewRedactorOptions openAIConfigs

// SetOpenAIAPIKeys sets API key and organization of options.
func (o *NewRedactorOptions) SetOpenAIAPIKeys(apiKey, organization string) *NewRedactorOptions {
	o.apiKey = &apiKey
	o.organization = &organization

	return o
}

// SetModel sets the model of options.
func (o *NewRedactorOptions) SetModel(model string) *NewRedactorOptions {
	o.model = &model

	return o
}

// SetVerbose sets the verbosity.
func (o *NewRedactorOptions) SetVerbose(verbose bool) *NewRedactorOptions {
	o.verbose = verbose

	return o
}

// NewRedactor returns a new redactor client with given `options`.
func NewRedactor(options *NewRedactorOptions) (client *Redactor, err error) {
	if options == nil {
		options = &NewRedactorOptions{}
	}

	if options.apiKey == nil || options.organization == nil {
		return nil, fmt.Errorf("no `apiKey` or `organization` was given")
	}

	if options.model == nil {
		model := defaultModel
		options.model = &model
	}

	openAIClient := openai.NewClient(*options.apiKey, *options.organization)
	openAIClient.Verbose = options.verbose

	client = &Redactor{
		openAI: &openAIConfigs{
			apiKey:       options.apiKey,
			organization: options.organization,
			model:        options.model,
			client:       openAIClient,
		},
	}

	var models openai.ModelsList
	if models, err = client.openAI.client.ListModels(); err != nil {
		return nil, fmt.Errorf("failed to list OpenAI models: %s", err)
	} else {
		if options.model == nil {
			return nil, fmt.Errorf("no `model` was given")
		}

		// check if given `model` exists in the models list
		for _, model := range models.Data {
			if model.ID == *options.model {
				return client, nil
			}
		}

		return nil, fmt.Errorf("no such model: %s in supported models list", *options.model)
	}
}

const fnNameRedact = "detect_private_or_sensitive_info"
const fnDescRedact = "This function detects private or sensitive information from the given text." // FIXME: improve this description

// Detect analyzes given `text` and returns detected private/sensitive information as a string array.
func (c *Redactor) Detect(text string) (detected []string, err error) {
	options := openai.ChatCompletionOptions{}.
		SetTemperature(chatCompletionTemperature).
		SetTools([]openai.ChatCompletionTool{
			openai.NewChatCompletionTool(fnNameRedact,
				fnDescRedact,
				openai.NewToolFunctionParameters().
					AddArrayPropertyWithDescription("detected", "string", "An array of detected private or sensitive information texts").
					SetRequiredParameters([]string{"detected"})),
		}).
		SetUser(userAgent)
	messages := []openai.ChatMessage{
		openai.NewChatAssistantMessage("You are an AI administrator who detects private or sensitive information from texts passed by the user."),
		openai.NewChatUserMessage(text),
	}

	var completed openai.ChatCompletion
	if completed, err = c.openAI.client.CreateChatCompletion(*c.openAI.model, messages, options); err == nil {
		if len(completed.Choices) > 0 {
			toolCalls := completed.Choices[0].Message.ToolCalls

			if len(toolCalls) > 0 {
				toolCall := toolCalls[0]
				if toolCall.Function.Name == fnNameRedact {
					var args struct {
						Detected []string `json:"detected"`
					}
					if err = toolCall.ArgumentsInto(&args); err == nil {
						return args.Detected, nil
					}
				} else {
					err = fmt.Errorf("returned tool call function name: `%s` differs from the requested one: `%s`", toolCall.Function.Name, fnNameRedact)
				}
			} else {
				err = fmt.Errorf("there was no tool call in chat completion choice from OpenAI")
			}
		} else {
			err = fmt.Errorf("there was no choice in chat completion from OpenAI")
		}
	}

	return nil, err
}

// Redact redacts given `text` by replacing detected strings with `to`.
func (c *Redactor) Redact(text, to string) (redacted string, err error) {
	return c.RedactFunc(text, func(_ string) string {
		return to
	})
}

// RedactFunc redacts given `text` by replacing detected strings with the result of function `fn`.
func (c *Redactor) RedactFunc(text string, fn func(in string) string) (redacted string, err error) {
	redacted = text

	var detected []string
	if detected, err = c.Detect(text); err == nil {
		if len(detected) <= 0 {
			return redacted, fmt.Errorf("there was no private/sensitive information in the given text")
		}

		for _, d := range detected {
			redacted = strings.ReplaceAll(redacted, d, fn(d))
		}
	}

	return redacted, err
}
