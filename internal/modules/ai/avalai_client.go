package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
)

type avalaiClient struct {
	client openai.Client
}

func newAvalaiClient(apiKey string) *avalaiClient {
	client := openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL("https://api.avalai.ir/v1"))
	return &avalaiClient{client: client}
}

func (c *avalaiClient) ask(dbContext string, question string) (string, error) {
	ctx := context.Background()

	// Build the system message with database context
	systemMessage := `You are a SQL query generator. Given a database schema and a natural language question, generate a valid SQL query.
Return ONLY the SQL query without any explanations, markdown formatting, or additional text.
If the question cannot be answered with the given schema, return an empty string.`

	// Build the user message with database context and question
	userMessage := fmt.Sprintf("Database Schema:\n%s\n\nQuestion: %s\n\nGenerate a SQL query:", dbContext, question)

	// Create the chat completion request
	chatCompletion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: "gpt-4o",
		Messages: []openai.ChatCompletionMessageParamUnion{
			{
				OfSystem: &openai.ChatCompletionSystemMessageParam{
					Content: openai.ChatCompletionSystemMessageParamContentUnion{
						OfString: param.Opt[string]{Value: systemMessage},
					},
				},
			},
			{
				OfUser: &openai.ChatCompletionUserMessageParam{
					Content: openai.ChatCompletionUserMessageParamContentUnion{
						OfString: param.Opt[string]{Value: userMessage},
					},
				},
			},
		},
		Temperature: param.Opt[float64]{Value: 0.3}, // Lower temperature for more deterministic SQL generation
	})

	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	// Extract the SQL query from the response
	if len(chatCompletion.Choices) == 0 {
		return "", fmt.Errorf("no choices in chat completion response")
	}

	sqlQuery := strings.TrimSpace(chatCompletion.Choices[0].Message.Content)

	// Remove markdown code blocks if present
	sqlQuery = strings.TrimPrefix(sqlQuery, "```sql")
	sqlQuery = strings.TrimPrefix(sqlQuery, "```")
	sqlQuery = strings.TrimSuffix(sqlQuery, "```")
	sqlQuery = strings.TrimSpace(sqlQuery)

	return sqlQuery, nil
}
