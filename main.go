package main

import (
    "bufio"
    "fmt"
    "github.com/sashabaranov/go-openai"
    log "github.com/sirupsen/logrus"
    "io"
    "os"
    "strings"
)

const CharacterPrompt = `Представь, что ты не чат-бот, а маленькая желтая 
  резиновая уточка, по имени Дакки. А я - твой друг Станислав. Ты будешь помогать мне в моих ежедневных задачах. Ты можешь отвечать на любой вопрос, 
  но если ты не уверен в ответе, то говоришь об этом. Ты, ни в коем 
  случае, не должен признаваться, что ты чат-бот или что ты используешь 
  нейросети. Ты не чат-бот, а маленькая желтая уточка. Тем не менее, ты заботишься о людях.`

var gpt *GPT
var history []openai.ChatCompletionMessage
var writer io.Writer = os.Stdout
var commander *Commander

func init() {
    log.SetFormatter(&log.TextFormatter{})
    log.SetLevel(log.DebugLevel)
}

func main() {
    ai := openai.NewClient(ensureEnv("OPENAI_TOKEN"))
    gpt = NewGPT(ai, openai.ChatCompletionRequest{
        Model:     openai.GPT3Dot5Turbo,
        MaxTokens: 1000,
    })

    history = append(history, openai.ChatCompletionMessage{
        Role:    openai.ChatMessageRoleSystem,
        Content: CharacterPrompt,
    })

    commander, err := NewCommander([]Command{
        NewCommandClearScreen(),
        NewCommandJiraGetTasks(),
    })
    if err != nil {
        log.Fatalf("failed to create commander: %v", err)
    }

    reader := bufio.NewScanner(os.Stdin)
    printPrompt()
    for reader.Scan() {
        text := cleanInput(reader.Text())
        if text == "" {
            printPrompt()
            continue
        }
        command, err := commander.ExtractCommand(text)
        if err != nil {
            printErrorToStdout(err)
            continue
        }
        log.Debugf("Command is `%v`", command)
        if command == nil {
            err := handleText(text)
            if err != nil {
                printErrorToStdout(err)
            }
        } else {
            output, err := command.Execute()
            if err != nil {
                printErrorToStdout(err)
            }
            if output != "" {
                fmt.Println(output)
            }
        }
        printPrompt()
    }
    // Print an additional line if we encountered an EOF character
    fmt.Println()
}

func printErrorToStdout(err error) {
    log.Errorf("Error: %v", err)
}

// handleText parses the given commands
func handleText(text string) error {
    history = append(history, openai.ChatCompletionMessage{
        Role:    openai.ChatMessageRoleUser,
        Content: text,
    })

    err := gpt.AskChatGPTAndStreamResponse(history, writer)
    if err != nil {
        return fmt.Errorf("failed to ask chat gpt: %w", err)
    }

    return nil
}

func limitChatCompletionMessages(chatCompletions []openai.ChatCompletionMessage, maxLength int) []openai.ChatCompletionMessage {
    var filteredCompletions []openai.ChatCompletionMessage
    totalContent := ""
    for i := len(chatCompletions) - 1; i >= 0; i-- {
        completion := chatCompletions[i]
        if len(totalContent)+len(completion.Content) > maxLength {
            break
        }
        totalContent += completion.Content
        filteredCompletions = append([]openai.ChatCompletionMessage{completion}, filteredCompletions...)
    }
    return filteredCompletions
}

// printPrompt displays the repl prompt at the start of each loop
func printPrompt() {
    fmt.Print("> ")
}

// cleanInput preprocesses input to the db repl
func cleanInput(text string) string {
    output := strings.TrimSpace(text)
    output = strings.ToLower(output)
    return output
}

func ensureEnv(key string) string {
    val := os.Getenv(key)
    if val == "" {
        log.Panicf("`%s` must be set.\n", key)
    }
    return val
}
