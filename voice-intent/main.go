package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

// ASR 请求结构体
type ASRRequest struct {
	Audio []byte `json:"audio"`
}

// LLM 请求结构体
type LLMRequest struct {
	Text string `json:"text"`
}

// 最终响应结构体
type Response struct {
	ASRText string `json:"asr_text"`
	Intent  string `json:"intent"`
}

// 调用 ASR 服务（模拟 WebSocket 调用）
func callASRService(audio []byte) (string, error) {
	// 替换为你的 ASR WebSocket 地址
	asrURL := url.URL{Scheme: "wss", Host: "dashscope.aliyuncs.com", Path: "/api-ws/v1/inference"}

	conn, _, err := websocket.DefaultDialer.Dial(asrURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("ASR 连接失败: %v", err)
	}
	defer conn.Close()

	// 发送音频数据
	if err := conn.WriteMessage(websocket.BinaryMessage, audio); err != nil {
		return "", fmt.Errorf("ASR 发送失败: %v", err)
	}

	// 接收识别结果
	_, message, err := conn.ReadMessage()
	if err != nil {
		return "", fmt.Errorf("ASR 接收失败: %v", err)
	}

	return string(message), nil
}

// 调用 LLM 服务（HTTP）
func callLLMService(text string) (string, error) {
	// 替换为你的 LLM API 地址
	llmURL := "https://dashscope.aliyuncs.com/api/v1/services/llm"

	reqBody, _ := json.Marshal(LLMRequest{Text: text})
	resp, err := http.Post(llmURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("LLM 调用失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Output string `json:"output"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("LLM 解析失败: %v", err)
	}

	return result.Output, nil
}

// 主处理函数
func processHandler(c echo.Context) error {
	// 1. 获取音频数据（假设通过 multipart/form-data 上传）
	file, err := c.FormFile("audio")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "需要上传音频文件"})
	}

	audioFile, _ := file.Open()
	defer audioFile.Close()
	audioData, _ := io.ReadAll(audioFile)

	// 2. 调用 ASR
	asrText, err := callASRService(audioData)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ASR 处理失败"})
	}

	// 3. 调用 LLM
	intent, err := callLLMService(asrText)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "意图识别失败"})
	}

	// 4. 返回结果
	return c.JSON(http.StatusOK, Response{
		ASRText: asrText,
		Intent:  intent,
	})
}

func main() {
	e := echo.New()
	e.POST("/process", processHandler)

	// 健康检查端点
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	log.Fatal(e.Start(":8080"))
}
