package main

import (
    "bufio"
    "io"
    "korchasa/little-duck/pkg/chat_gpt"
    "korchasa/little-duck/pkg/utils"
    "log"
    "os"
)

func main() {
    var writer io.Writer = os.Stdout
    modelPath := os.Args[1]

    model, err := chat_gpt.LoadModel(modelPath)
    if err != nil {
        log.Panicf("failed to load model: %v", err)
    }

    reader := bufio.NewScanner(os.Stdin)
    utils.PrintPrompt()
    for reader.Scan() {
        text := utils.ClearInput(reader.Text())
        if text == "" {
            utils.PrintPrompt()
            continue
        }
        if text == "exit" {
            break
        }
        err = model.AskChatGPTAndStreamResponse(text, writer)
        if err != nil {
            log.Panicf("failed to ask chat gpt: %v", err)
        }
    }
}
