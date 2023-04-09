package chat_gpt

import (
    "context"
    "errors"
    "fmt"
    "github.com/sashabaranov/go-openai"
    log "github.com/sirupsen/logrus"
    "gopkg.in/yaml.v3"
    "io"
    "korchasa/little-duck/pkg/utils"
    "os"
)

var (
    openAIClient             *openai.Client
    MaxChatMessagesLengthSum = 4000
)

func init() {
    openAIClient = openai.NewClient(utils.EnsureEnv("OPENAI_API_KEY"))
}

type Model struct {
    req openai.ChatCompletionRequest
}

func LoadModel(path string) (*Model, error) {
    sb, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read model file: %w", err)
    }
    var req openai.ChatCompletionRequest
    err = yaml.Unmarshal(sb, &req)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal model: %w", err)
    }

    return &Model{
        req: req,
    }, nil
}

func (m *Model) AskChatGPTAndStreamResponse(text string, wr io.Writer) error {
    m.AddUserMessage(text)
    req := m.req
    req.Messages = limitChatCompletionMessages(req.Messages, MaxChatMessagesLengthSum)
    req.Stream = true
    stream, err := openAIClient.CreateChatCompletionStream(context.TODO(), req)
    if err != nil {
        return fmt.Errorf("failed to create chat completion stream: %w", err)
    }
    defer stream.Close()

    var fullResponse string
    for {
        response, err := stream.Recv()
        if errors.Is(err, io.EOF) {
            m.req.Messages = append(req.Messages, openai.ChatCompletionMessage{
                Role:    openai.ChatMessageRoleAssistant,
                Content: fullResponse,
            })
            _, err = wr.Write([]byte("\n"))
            if err != nil {
                return fmt.Errorf("failed to write chat completion message: %w", err)
            }
            return nil
        }
        if err != nil {
            return fmt.Errorf("failed to receive chat completion message: %w", err)
        }
        _, err = wr.Write([]byte(response.Choices[0].Delta.Content))
        if err != nil {
            return fmt.Errorf("failed to write chat completion message: %w", err)
        }
    }
}

func (m *Model) AskChatGPTAndGetResponse(text string) (string, error) {
    m.AddUserMessage(text)
    req := m.req
    req.Messages = limitChatCompletionMessages(req.Messages, MaxChatMessagesLengthSum)
    resp, err := openAIClient.CreateChatCompletion(context.TODO(), req)
    if err != nil {
        return "", fmt.Errorf("failed to create chat completion request: %w", err)
    }
    for i, ch := range resp.Choices {
        log.Debugf("Choises: index=%d name=%v content=%v", i, ch.Message.Name, ch.Message.Content)
    }
    respText := resp.Choices[0].Message.Content
    m.req.Messages = append(req.Messages, openai.ChatCompletionMessage{
        Role:    openai.ChatMessageRoleAssistant,
        Content: respText,
    })
    return respText, nil
}

func (m *Model) AddUserMessage(text string) {
    m.req.Messages = append(m.req.Messages, openai.ChatCompletionMessage{
        Role:    openai.ChatMessageRoleUser,
        Content: text,
    })
}

func limitChatCompletionMessages(chatCompletions []openai.ChatCompletionMessage, maxLength int) []openai.ChatCompletionMessage {
    var filteredCompletions []openai.ChatCompletionMessage
    totalContent := ""
    for i := len(chatCompletions) - 1; i >= 0; i-- {
        completion := chatCompletions[i]
        if len(totalContent)+len(completion.Content) > maxLength {
            break
        }
        totalContent += completion.Content
        filteredCompletions = append([]openai.ChatCompletionMessage{completion}, filteredCompletions...)
    }
    return filteredCompletions
}
