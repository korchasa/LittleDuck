package main

import (
    "os"
    "os/exec"
)

type CommandClearScreen struct {
}

func NewCommandClearScreen() *CommandClearScreen {
    return &CommandClearScreen{}
}

func (c *CommandClearScreen) Name() string {
    return "clearScreen"
}

func (c *CommandClearScreen) Prompts() []PromptExample {
    return []PromptExample{
        {
            "clear",
            CommandSpec{"clearScreen", []string{}},
        },
        {
            "очисти экран",
            CommandSpec{"clearScreen", []string{}},
        },
    }
}

func (c *CommandClearScreen) Execute(args []string) (string, error) {
    cmd := exec.Command("clear")
    cmd.Stdout = os.Stdout
    return "", cmd.Run()
}
