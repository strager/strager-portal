package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	launchd "github.com/bored-engineer/go-launchd"
	sse "github.com/tmaxmax/go-sse"
	"html"
	"io"
	"iter"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type appConfig struct {
	GroqAPIKey string `toml:"groq_api_key"`
}

var config appConfig = appConfig{}

func main() {
	var err error

	var exePath string
	exePath, err = os.Executable()
	if err != nil {
		slog.Error("failed to get path to executable", slog.Any("err", err))
		os.Exit(1)
	}
	var configPath string = filepath.Join(filepath.Dir(exePath), "config.toml")

	_, err = toml.DecodeFile(configPath, &config)
	if err != nil {
		slog.Error("failed to load config file", slog.Any("err", err), slog.Any("configPath", configPath))
		os.Exit(1)
	}
	if config.GroqAPIKey == "" {
		slog.Error("missing groq_api_key in config file", slog.Any("configPath", configPath))
		os.Exit(1)
	}

	slog.Info("starting service")
	http.HandleFunc("/", handleRequest)

	var serverSocket net.Listener
	serverSocket, err = launchd.Activate("Listeners")
	if err != nil {
		if errors.Is(err, syscall.ESRCH) {
			slog.Debug("not launched by launchd")
			serverSocket = nil
		} else {
			slog.Error("failed to get socket from launchd", slog.Any("error", err))
			os.Exit(1)
		}
	}

	if serverSocket == nil {
		err = http.ListenAndServe("localhost:69", nil)
	} else {
		err = http.Serve(serverSocket, nil)
	}
	if err != nil {
		slog.Error("failed to start HTTP server", slog.Any("error", err))
		os.Exit(1)
	}
}

func handleRequest(response http.ResponseWriter, request *http.Request) {
	var query string = request.URL.Query().Get("q")
	var tokens []string = strings.Fields(query)

	var bangTokenIndex int = -1
	if len(tokens) > 0 {
		if strings.HasPrefix(tokens[0], "!") {
			bangTokenIndex = 0
		} else if strings.HasPrefix(tokens[len(tokens)-1], "!") {
			bangTokenIndex = len(tokens) - 1
		}
	}

	if bangTokenIndex != -1 {
		var queryWithoutBang string = ""
		var i int
		for i = range tokens {
			if i == bangTokenIndex {
				continue
			}
			if queryWithoutBang != "" {
				queryWithoutBang += " "
			}
			queryWithoutBang += tokens[i]
		}
		redirectToBang(tokens[bangTokenIndex], response, request, queryWithoutBang)
		return
	}

	if strings.HasSuffix(query, "?") {
		showAIConversation(response, query)
		return
	}

	redirectToBang("", response, request, query)
}

// bang is either "!w" (for example) or "!" or no bang.
func redirectToBang(bang string, response http.ResponseWriter, request *http.Request, queryWithoutBang string) {
	var handler bangHandler = bangHandlers[bang]
	if handler == nil {
		handler = bangHandlers[""]
		if handler == nil {
			response.WriteHeader(http.StatusNotFound)
			_, _ = io.WriteString(response, "default handler not found")
			return
		}
		handler(response, request, bang+" "+queryWithoutBang)
	}
	handler(response, request, queryWithoutBang)
}

func showAIConversation(response http.ResponseWriter, query string) {
	var err error

	// TODO(strager): Remove trailing "?".
	var conversation []string = []string{query}

	var flush http.Flusher
	flush, _ = response.(http.Flusher)

	var wroteHeader bool = false
	var writeHeaderIfNeeded = func() {
		if wroteHeader {
			return
		}
		response.Header().Set("Content-Type", "text/html; charset=utf-8")
		response.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(response, `<!DOCTYPE html>
<meta charset="utf-8">
<meta name="color-scheme" content="dark" />

<style>
body {
  font-family: sans-serif;
}

#chat .user, #chat .assistant {
  padding: 1rem;
}

#chat .assistant {
  white-space: pre;
}
</style>

<div id="chat">
<div class="user">`)
		_, _ = io.WriteString(response, html.EscapeString(query))
		_, _ = io.WriteString(response, `</div>
<div class="assistant">`)
		wroteHeader = true
	}
	var chunk AICompletionChunk
	for chunk, err = range streamAI(conversation) {
		if err != nil {
			slog.Error("error streaming AI completion", slog.Any("error", err))
			// TODO(strager): Show error to user.
			response.WriteHeader(http.StatusInternalServerError)
			wroteHeader = true
			return
		}

		var content string = ""
		if len(chunk.Choices) > 0 {
			content = chunk.Choices[0].Delta.Content
		}
		if content != "" {
			writeHeaderIfNeeded()
			_, _ = io.WriteString(response, html.EscapeString(content))
			if flush != nil {
				flush.Flush()
			}
		}
	}

	writeHeaderIfNeeded()
	_, _ = io.WriteString(response, `</div>
`)
}

type AICompletionChunkDelta struct {
	Content string `json:"content"`
}

type AICompletionChunkChoice struct {
	Delta AICompletionChunkDelta `json:"delta"`
}

type AICompletionChunk struct {
	Choices []AICompletionChunkChoice `json:"choices"`
}

func streamAI(conversation []string) iter.Seq2[AICompletionChunk, error] {
	return func(yield func(AICompletionChunk, error) bool) {
		var err error

		var aiResponse *http.Response
		aiResponse, err = requestAI(conversation)
		if err != nil {
			yield(AICompletionChunk{}, err)
			return
		}
		defer aiResponse.Body.Close()

		var event sse.Event
		for event, err = range sse.Read(aiResponse.Body, &sse.ReadConfig{}) {
			if event.Data == "[DONE]" {
				break
			}
			var chunk AICompletionChunk
			err = json.Unmarshal([]byte(event.Data), &chunk)
			if err != nil {
				if !yield(chunk, fmt.Errorf("decoding chunk: %w", err)) {
					break
				}
				continue
			}
			if !yield(chunk, nil) {
				break
			}
		}
	}
}

func requestAI(conversation []string) (*http.Response, error) {
	var err error

	type Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type RequestBody struct {
		Messages            []Message `json:"messages"`
		Model               string    `json:"model"`
		MaxCompletionTokens int       `json:"max_completion_tokens"`
		Stream              bool      `json:"stream"`
	}

	var messages []Message = []Message{
		Message{
			Role:    "system",
			Content: "You are an assistant helping an experienced software engineer. The engineer is requesting information. Please provide the information requested in a concise manner without headings or unnecessary explanation. If appropriate, show a short code example in the language mentioned. Keep commentary to a minimum. Unless requested, do not include error handling code.",
		},
	}
	var i int
	for i = range conversation {
		var role string
		if i%2 == 0 {
			role = "user"
		} else {
			role = "assistant"
		}
		messages = append(messages, Message{Role: role, Content: conversation[i]})
	}

	var body []byte
	body, err = json.Marshal(RequestBody{
		Messages:            messages,
		Model:               "llama-3.3-70b-versatile",
		MaxCompletionTokens: 1000,
		Stream:              true,
	})
	if err != nil {
		return nil, err
	}

	var request *http.Request
	request, err = http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+config.GroqAPIKey)

	var response *http.Response
	var client *http.Client = &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		return nil, err
	}
	return response, nil
}
