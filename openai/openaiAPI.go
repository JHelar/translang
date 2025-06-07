package openai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type OpenaiClient struct {
	openai.Client
	translatorParam openai.ChatCompletionNewParams
}

type Translation struct {
	Source  string `json:"source" jsonschema_description:"The source text that the translations came from"`
	English string `json:"english" jsonschema_description:"The english translation"`
	Swedish string `json:"swedish" jsonschema_description:"The swedish translation"`
	Finnish string `json:"finnish" jsonschema_description:"The finnish translation"`
	CopyKey string `json:"copyKey" jsonschema_description:"A suggested copy key to be used as a reference to the translations"`
}

func (translation Translation) String() string {
	return fmt.Sprintf("Key: %v\nSource: %v\nEN: %v\nSV: %v\nFI: %v", translation.CopyKey, translation.Source, translation.English, translation.Swedish, translation.Finnish)
}

func GenerateScehma[T any]() interface{} {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var value T
	schema := reflector.Reflect(value)
	return schema
}

var TranslationsResponseSchema = GenerateScehma[Translation]()

func NewClient(openaiAPIKey string) OpenaiClient {
	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "translations",
		Description: openai.String("The languages that will be translated into"),
		Schema:      TranslationsResponseSchema,
		Strict:      openai.Bool(true),
	}

	translatorParam := openai.ChatCompletionNewParams{
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: schemaParam,
			},
		},
		Model: openai.ChatModelGPT4o2024_08_06,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(TRANSLATION_SYSTEM_PROMPT),
		},
	}

	return OpenaiClient{
		openai.NewClient(
			option.WithAPIKey(openaiAPIKey),
		),
		translatorParam,
	}

}

const TRANSLATION_SYSTEM_PROMPT = `
	You are the leading copy translator world champion. 
	You will translate any given string to copy that will be used in production products.
	Your only purpose is to translate, you do not give any explanation or other information.
	You only respond with the translations and a suggested copy key.
	A copy key is a string that will be used in the product to reference the translations. Derive the key from the english translation of the source text, keep it small, not verbose, format it using snake uppercase.
	You do not diverge the translation from the given source text.
	From now on any text that I give you, you will only respond with the translation and copy key of that text nothing more nothing less.
`

func (client *OpenaiClient) Translate(text string) Translation {
	client.translatorParam.Messages = append(client.translatorParam.Messages, openai.UserMessage(text))

	chat, _ := client.Chat.Completions.New(
		context.TODO(),
		client.translatorParam,
	)

	client.translatorParam.Messages = client.translatorParam.Messages[:1]

	var translation Translation
	_ = json.Unmarshal([]byte(chat.Choices[0].Message.Content), &translation)

	return translation
}
