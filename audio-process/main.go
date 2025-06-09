package main

import (
  "github.com/alibaba/higress/plugins/wasm-go/pkg/wrapper"
  "github.com/higress-group/proxy-wasm-go-sdk/proxywasm/types"
  "github.com/tidwall/gjson"
)

func main() {
  wrapper.SetCtx(
    // 插件名称
    "easy-logger",
    // 设置自定义函数解析插件配置
    wrapper.ParseConfigBy(parseConfig),
    // 设置自定义函数处理请求头
    wrapper.ProcessRequestHeadersBy(onHttpRequestHeaders),
    // 设置自定义函数处理请求体
    wrapper.ProcessRequestBodyBy(onHttpRequestBody),
    // 设置自定义函数处理响应头
    wrapper.ProcessResponseHeadersBy(onHttpResponseHeaders),
    // 设置自定义函数处理响应体
    wrapper.ProcessResponseBodyBy(onHttpResponseBody),
    // 设置自定义函数处理流式请求体
    //wrapper.ProcessStreamingRequestBodyBy(onHttpStreamingRequestBody),
    // 设置自定义函数处理流式响应体
    //wrapper.ProcessStreamingResponseBodyBy(onHttpStreamingResponseBody),
  )
}

// 自定义插件配置
type LoggerConfig struct {
  // 是否打印请求
  request bool
  // 是否打印响应
  response bool
  // 打印响应状态码，* 表示打印所有状态响应，500,502,503 表示打印 HTTP 500、502、503 状态响应，默认是 *
  responseStatusCodes string
}

func parseConfig(json gjson.Result, config *LoggerConfig, log wrapper.Log) error {
  log.Debugf("parseConfig()")
  config.request = json.Get("request").Bool()
  config.response = json.Get("response").Bool()
  config.responseStatusCodes = json.Get("responseStatusCodes").String()
  if config.responseStatusCodes == "" {
    config.responseStatusCodes = "*"
  }
  log.Debugf("parse config:%v", config)
  return nil
}

func onHttpRequestHeaders(ctx wrapper.HttpContext, config LoggerConfig, log wrapper.Log) types.Action {
  log.Debugf("onHttpRequestHeaders()")
  return types.ActionContinue
}

func onHttpRequestBody(ctx wrapper.HttpContext, config LoggerConfig, body []byte, log wrapper.Log) types.Action {
  log.Debugf("onHttpRequestBody()")
  return types.ActionContinue
}

func onHttpResponseBody(ctx wrapper.HttpContext, config LoggerConfig, body []byte, log wrapper.Log) types.Action {
  log.Debugf("onHttpResponseBody()")
  return types.ActionContinue
}

func onHttpResponseHeaders(ctx wrapper.HttpContext, config LoggerConfig, log wrapper.Log) types.Action {
  log.Debugf("onHttpResponseHeaders()")
  return types.ActionContinue
}