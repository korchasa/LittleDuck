package main

import (
    "os"
    "os/exec"
)

type CommandChangeLogLevel struct {
    args []string
}

func NewCommandChangeLogLevel() *CommandChangeLogLevel {
    return &CommandChangeLogLevel{}
}

func (c *CommandChangeLogLevel) Name() string {
    return "clearScreen"
}

func (c *CommandChangeLogLevel) Execute(_ []string) (string, error) {
    cmd := exec.Command("clear")
    cmd.Stdout = os.Stdout
    return "", cmd.Run()
}
