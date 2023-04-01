package main

import (
    "os"
    "os/exec"
)

type CommandClearScreen struct {
    args []string
}

func NewCommandClearScreen() *CommandClearScreen {
    return &CommandClearScreen{}
}

func (c *CommandClearScreen) Name() string {
    return "clearScreen"
}

func (c *CommandClearScreen) Prompts() []Prompt {
    return []Prompt{
        {
            "clear",
            ExecSpec{"clearScreen", []string{}},
        },
        {
            "очисти экран",
            ExecSpec{"clearScreen", []string{}},
        },
    }
}

func (c *CommandClearScreen) SetArguments(args []string) {
    c.args = args
}

func (c *CommandClearScreen) Execute() (string, error) {
    cmd := exec.Command("clear")
    cmd.Stdout = os.Stdout
    return "", cmd.Run()
}
