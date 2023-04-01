package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "github.com/sashabaranov/go-openai"
    "text/template"
)

type Prompt struct {
    Input  string
    Output ExecSpec
}

type ExecSpec struct {
    Name      string   `json:"command"`
    Arguments []string `json:"arguments"`
}

type Command interface {
    Name() string
    SetArguments(args []string)
    Prompts() []Prompt
    Execute() (string, error)
}

type Commander struct {
    commands []Command
    template *template.Template
}

func NewCommander(commands []Command) (*Commander, error) {
    tpl, err := template.New("commander").Parse(`Turn a user message into a command description, in json format:
{"command": "<command name>", "arguments": ["<argument 1>", "<argument 2>", ...]}

The arguments must be strings. The command name can be one of:
{{ range . }}
 - {{ .Name }}
{{- end }}

Examples:
{{ range . }}
    {{ range .Prompts }}
Example: {{ .Input }}
Output: {"command": "{{ .Output.Name }}", "arguments": [{{ range $index, $element := .Output.Arguments }}{{if $index}}, {{end}}"{{ . }}"{{ end }}]}
    {{ end }}
{{- end }}

If you can't determine the type of command, answer with an empty string.
`)
    if err != nil {
        return nil, fmt.Errorf("failed to parse commander template: %w", err)
    }
    return &Commander{
        commands: commands,
        template: tpl,
    }, nil
}

func (c *Commander) ExtractCommand(text string) (Command, error) {
    var prompt bytes.Buffer
    err := c.template.Execute(&prompt, c.commands)
    if err != nil {
        return nil, fmt.Errorf("failed to execute template: %w", err)
    }

    resp, err := gpt.AskChatGPTAndGetResponse([]openai.ChatCompletionMessage{
        {
            Role:    openai.ChatMessageRoleSystem,
            Content: prompt.String(),
        },
        {
            Role:    openai.ChatMessageRoleUser,
            Content: fmt.Sprintf("Example: %s\nOutput: ", text),
        },
    })
    if err != nil {
        return nil, fmt.Errorf("failed to ask chat gpt: %w", err)
    }

    spec := &ExecSpec{}
    err = json.Unmarshal([]byte(resp), spec)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal command spec: %w", err)
    }

    command := c.findCommand(spec.Name)
    if command == nil {
        return nil, nil
    }
    command.SetArguments(spec.Arguments)
    return command, nil
}

func (c *Commander) findCommand(name string) Command {
    for _, cmd := range c.commands {
        if cmd.Name() == name {
            return cmd
        }
    }
    return nil
}
