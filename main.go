package main

import (
    "bufio"
    "fmt"
    log "github.com/sirupsen/logrus"
    "io"
    "korchasa/little-duck/pkg/chat_gpt"
    "korchasa/little-duck/pkg/utils"
    "os"
)

const mainModelPath = "./models/general.yaml"

var writer io.Writer = os.Stdout

func init() {
    log.SetFormatter(&log.TextFormatter{})
    log.SetLevel(log.DebugLevel)
}

func main() {
    commander, err := NewCommandSelector([]Command{
        NewCommandClearScreen(),
        NewCommandJiraGetTasks(),
    })
    if err != nil {
        log.Fatalf("failed to create commander: %v", err)
    }

    mainModel, err := chat_gpt.LoadModel(mainModelPath)
    if err != nil {
        log.Panicf("failed to load main model: %v", err)
    }

    reader := bufio.NewScanner(os.Stdin)
    utils.PrintPrompt()
    for reader.Scan() {
        text := utils.ClearInput(reader.Text())
        if text == "" {
            utils.PrintPrompt()
            continue
        }
        command, query, err := commander.ExtractCommand(text)
        if err != nil {
            utils.PrintErrorToStdout(err)
            continue
        }
        if command == nil {
            log.Debugf("Not a command text: %s", text)
            err := handleText(mainModel, text)
            if err != nil {
                utils.PrintErrorToStdout(err)
            }
        } else {
            log.Debugf("Command: %s(%s)", command.Name(), query)
            output, err := command.Execute(query)
            if err != nil {
                utils.PrintErrorToStdout(err)
            }
            if output != "" {
                fmt.Println(output)
            }
        }
        utils.PrintPrompt()
    }
    // Print an additional line if we encountered an EOF character
    fmt.Println()
}

// handleText parses the given commands
func handleText(model *chat_gpt.Model, text string) error {
    err := model.AskChatGPTAndStreamResponse(text, writer)
    if err != nil {
        return fmt.Errorf("failed to ask chat gpt: %w", err)
    }

    return nil
}

func ensureEnv(key string) string {
    val := os.Getenv(key)
    if val == "" {
        log.Panicf("`%s` must be set.\n", key)
    }
    return val
}
