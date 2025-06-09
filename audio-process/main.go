package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/alibaba/higress/plugins/wasm-go/pkg/wrapper"
	"github.com/gorilla/websocket"
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm"
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm/types"
	"github.com/tidwall/gjson"
)

func main() {
	wrapper.SetCtx(
		"audio-intent-processor",
		wrapper.ParseConfigBy(parseConfig),
		wrapper.ProcessRequestBodyBy(onHttpRequestBody),
	)
}

type PluginConfig struct {
	ASRServiceURL string
	LLMServiceURL string
	ASRAPIKey     string
	LLMAPIKey     string
}

func parseConfig(json gjson.Result, config *PluginConfig, log wrapper.Log) error {
	config.ASRServiceURL = json.Get("asrServiceURL").String()
	if config.ASRServiceURL == "" {
		return fmt.Errorf("missing asrServiceURL in config")
	}
	config.LLMServiceURL = json.Get("llmServiceURL").String()
	if config.LLMServiceURL == "" {
		return fmt.Errorf("missing llmServiceURL in config")
	}
	config.ASRAPIKey = json.Get("asrAPIKey").String()
	if config.ASRAPIKey == "" {
		return fmt.Errorf("missing asrAPIKey in config")
	}
	config.LLMAPIKey = json.Get("llmAPIKey").String()
	if config.LLMAPIKey == "" {
		return fmt.Errorf("missing llmAPIKey in config")
	}
	return nil
}

func onHttpRequestBody(ctx wrapper.HttpContext, config PluginConfig, body []byte, log wrapper.Log) types.Action {
	// Step 1: Send audio to ASR service
	asrResponse, err := sendToASRService(config.ASRServiceURL, body, log)
	if err != nil {
		log.Errorf("ASR service error: %v", err)
		proxywasm.SendHttpResponse(http.StatusInternalServerError, nil, []byte("ASR service failed"), -1)
		return types.ActionPause
	}

	// Step 2: Send text to LLM service
	llmResponse, err := sendToLLMService(config.LLMServiceURL, asrResponse, config, log)
	if err != nil {
		log.Errorf("LLM service error: %v", err)
		proxywasm.SendHttpResponse(http.StatusInternalServerError, nil, []byte("LLM service failed"), -1)
		return types.ActionPause
	}

	// Step 3: Return the result to the client
	proxywasm.SendHttpResponse(http.StatusOK, nil, []byte(llmResponse), -1)
	return types.ActionPause
}

func sendToASRService(url string, audio []byte, log wrapper.Log) (string, error) {
	// Establish a WebSocket connection to the ASR service
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Errorf("Failed to connect to ASR WebSocket: %v", err)
		return "", err
	}
	defer conn.Close()

	// Send the audio data over the WebSocket
	err = conn.WriteMessage(websocket.BinaryMessage, audio)
	if err != nil {
		log.Errorf("Failed to send audio to ASR WebSocket: %v", err)
		return "", err
	}

	// Read the response from the ASR service
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Errorf("Failed to read response from ASR WebSocket: %v", err)
		return "", err
	}

	return string(message), nil
}

func sendToLLMService(url string, text string, config PluginConfig, log wrapper.Log) (string, error) {
	// Prepare the OpenAI API request payload
	requestBody := map[string]interface{}{
		"model": "gpt-4",
		"messages": []map[string]string{
			{"role": "user", "content": text},
		},
	}

	// Serialize the request body to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Errorf("Failed to serialize LLM request body: %v", err)
		return "", err
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Errorf("Failed to create LLM request: %v", err)
		return "", err
	}

	// Add headers, including the API key
	req.Header.Set("Authorization", "Bearer "+config.LLMAPIKey)
	req.Header.Set("Content-Type", "application/json")

	// Send the request using an HTTP client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Failed to send LLM request: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failed to read LLM response: %v", err)
		return "", err
	}

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		log.Errorf("LLM service returned status %d: %s", resp.StatusCode, string(responseBody))
		return "", fmt.Errorf("LLM service error: %s", string(responseBody))
	}

	return string(responseBody), nil
}