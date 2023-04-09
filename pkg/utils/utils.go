package utils

import (
    "fmt"
    log "github.com/sirupsen/logrus"
    "os"
    "strings"
)

func PrintErrorToStdout(err error) {
    log.Errorf("Error: %v", err)
}

func EnsureEnv(key string) string {
    val := os.Getenv(key)
    if val == "" {
        log.Panicf("`%s` must be set.\n", key)
    }
    return val
}

// PrintPrompt displays the repl prompt at the start of each loop
func PrintPrompt() {
    fmt.Print("> ")
}

func ClearInput(text string) string {
    output := strings.TrimSpace(text)
    output = strings.ToLower(output)
    return output
}
