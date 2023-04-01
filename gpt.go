package main

import (
    "context"
    "errors"
    "fmt"
    "github.com/sashabaranov/go-openai"
    log "github.com/sirupsen/logrus"
    "io"
)

const MaxChatMessagesLengthSum = 4000

type GPT struct {
    ai       *openai.Client
    template openai.ChatCompletionRequest
}

func NewGPT(ai *openai.Client, template openai.ChatCompletionRequest) *GPT {
    return &GPT{
        ai:       ai,
        template: template,
    }
}

func (g *GPT) AskChatGPTAndStreamResponse(
    history []openai.ChatCompletionMessage,
    wr io.Writer) error {

    req := g.template
    req.Messages = limitChatCompletionMessages(history, MaxChatMessagesLengthSum)
    req.Stream = true
    stream, err := g.ai.CreateChatCompletionStream(context.TODO(), req)
    if err != nil {
        return fmt.Errorf("failed to create chat completion stream: %w", err)
    }
    defer stream.Close()

    var fullResponse string
    for {
        response, err := stream.Recv()
        if errors.Is(err, io.EOF) {
            history = append(history, openai.ChatCompletionMessage{
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

func (g *GPT) AskChatGPTAndGetResponse(history []openai.ChatCompletionMessage) (string, error) {
    req := g.template
    req.Messages = limitChatCompletionMessages(history, MaxChatMessagesLengthSum)
    resp, err := g.ai.CreateChatCompletion(context.TODO(), req)
    if err != nil {
        return "", fmt.Errorf("failed to create chat completion request: %w", err)
    }
    for i, ch := range resp.Choices {
        log.Debugf("Choises: index=%d name=%v content=%v", i, ch.Message.Name, ch.Message.Content)
    }
    return resp.Choices[0].Message.Content, nil
}
