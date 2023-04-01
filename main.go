package main

import (
    "bufio"
    "context"
    "errors"
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
const MaxChatMessagesLengthSum = 4000

var ai *openai.Client
var aiMessageTemplate = openai.ChatCompletionRequest{
    Model:     openai.GPT3Dot5Turbo,
    MaxTokens: 1000,
}
var history []openai.ChatCompletionMessage
var commands []Command = []Command{
    NewCommandClearScreen(),
    NewCommandJiraGetTasks(),
}

func init() {
    log.SetFormatter(&log.TextFormatter{})
    log.SetLevel(log.DebugLevel)
}

func main() {
    ai = openai.NewClient(ensureEnv("OPENAI_TOKEN"))
    history = append(history, openai.ChatCompletionMessage{
        Role:    openai.ChatMessageRoleSystem,
        Content: CharacterPrompt,
    })

    reader := bufio.NewScanner(os.Stdin)
    printPrompt()
    for reader.Scan() {
        text := cleanInput(reader.Text())
        if text == "" {
            printPrompt()
            continue
        }
        comm, err := extractCommand(text)
        if err != nil {
            printErrorToStdout(err)
            continue
        }
        log.Debugf("Command is `%v`", comm)
        if comm == nil {
            err := handleText(text)
            if err != nil {
                printErrorToStdout(err)
            }
        } else {
            for _, command := range commands {
                if command.Name() == comm.Name {
                    output, err := command.Execute(comm.Arguments)
                    if err != nil {
                        printErrorToStdout(err)
                    }
                    if output != "" {
                        fmt.Println(output)
                    }
                }
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
    cl, err := extractCommand(text)
    if err != nil {
        return fmt.Errorf("failed to clusterize text: %w", err)
    }
    log.Infof("Cluster is `%v`", cl)

    history = append(history, openai.ChatCompletionMessage{
        Role:    openai.ChatMessageRoleUser,
        Content: text,
    })
    req := aiMessageTemplate
    req.Messages = limitChatCompletionMessages(history, MaxChatMessagesLengthSum)
    req.Stream = true
    stream, err := ai.CreateChatCompletionStream(context.TODO(), req)
    if err != nil {
        return fmt.Errorf("failed to create chat completion stream: %w", err)
    }
    defer stream.Close()

    var fullResponse string
    for {
        response, err := stream.Recv()
        if errors.Is(err, io.EOF) {
            history = append(history, openai.ChatCompletionMessage{
                Role:    openai.ChatMessageRoleAssistant,
                Content: fullResponse,
            })
            fmt.Println()
            return nil
        }
        if err != nil {
            return fmt.Errorf("failed to receive chat completion message: %w", err)
        }
        fmt.Print(response.Choices[0].Delta.Content)
    }
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
