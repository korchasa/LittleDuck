package main

import (
    "encoding/json"
    "fmt"
    "korchasa/little-duck/pkg/chat_gpt"
)

const (
    CommandSelectorModelPath = "./models/command-selector.yaml"
)

type ExecSpec struct {
    Name  string `json:"command"`
    Query string `json:"query"`
}

type Command interface {
    Name() string
    Execute(query string) (string, error)
}

type CommandSelector struct {
    commands []Command
    model    *chat_gpt.Model
}

func NewCommandSelector(commands []Command) (*CommandSelector, error) {
    m, err := chat_gpt.LoadModel(CommandSelectorModelPath)
    if err != nil {
        return nil, fmt.Errorf("failed to load model `%s`: %w", CommandSelectorModelPath, err)
    }
    return &CommandSelector{
        commands: commands,
        model:    m,
    }, nil
}

func (c *CommandSelector) ExtractCommand(text string) (comm Command, query string, err error) {
    resp, err := c.model.AskChatGPTAndGetResponse(text)
    if err != nil {
        return nil, "", fmt.Errorf("failed to ask chat gpt: %w", err)
    }

    spec := &ExecSpec{}
    err = json.Unmarshal([]byte(resp), spec)
    if err != nil {
        return nil, "", fmt.Errorf("failed to unmarshal command spec: %w", err)
    }

    return c.findCommand(spec.Name), spec.Query, nil
}

func (c *CommandSelector) findCommand(name string) Command {
    for _, cmd := range c.commands {
        if cmd.Name() == name {
            return cmd
        }
    }
    return nil
}

func (c *CommandSelector) ExecuteSpec(spec ExecSpec) (string, error) {
    command := c.findCommand(spec.Name)
    if command == nil {
        return "", fmt.Errorf("command %s not found", spec.Name)
    }
    return command.Execute(spec.Query)
}
