package main

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/sashabaranov/go-openai"
    log "github.com/sirupsen/logrus"
)

type PromptExample struct {
    Input  string
    Output CommandSpec
}

type CommandSpec struct {
    Name      string   `json:"command"`
    Arguments []string `json:"arguments"`
}

type Command interface {
    Name() string
    Prompts() []PromptExample
    Execute(args []string) (string, error)
}

const ClusterPrePrompt = `Преврати сообщение пользователя в описание команды, в формате json:
{
  "command": "<название команды>",
  "arguments": [
    "<аргумент команды 1>", "<аргумент команды 2>", ...
  ]
}

Аргументы должны быть строками. Название команды может быть одним из:
`

func extractCommand(text string) (*CommandSpec, error) {
    names := ""
    examples := ""
    for _, m := range commands {
        names += fmt.Sprintf("- %s\n", m.Name())
        for _, p := range m.Prompts() {
            out, err := json.Marshal(p.Output)
            if err != nil {
                return nil, fmt.Errorf("failed to marshal output: %w", err)
            }
            examples += fmt.Sprintf(
                "Example: %s\nOutput: %s\n",
                p.Input,
                out,
            )
        }
    }
    prompt := fmt.Sprintf("%s%s%s", ClusterPrePrompt, names, examples)
    log.Debugf("Prompt: %s", prompt)

    req := aiMessageTemplate
    req.Messages = []openai.ChatCompletionMessage{
        {
            Role:    openai.ChatMessageRoleSystem,
            Content: prompt,
        },
        {
            Role:    openai.ChatMessageRoleUser,
            Content: fmt.Sprintf("Example: %s\nOutput: ", text),
        },
    }
    aiResp, err := ai.CreateChatCompletion(context.TODO(), req)
    if err != nil {
        return nil, fmt.Errorf("failed to create chat completion: %w", err)
    }
    for i, ch := range aiResp.Choices {
        log.Debugf("Choises: index=%d name=%v content=%v", i, ch.Message.Name, ch.Message.Content)
    }

    spec := &CommandSpec{}
    err = json.Unmarshal([]byte(aiResp.Choices[0].Message.Content), spec)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal command spec: %w", err)
    }
    return spec, nil
}
