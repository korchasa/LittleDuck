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

func (c *CommandClearScreen) Execute(_ string) (string, error) {
    cmd := exec.Command("clear")
    cmd.Stdout = os.Stdout
    return "", cmd.Run()
}
