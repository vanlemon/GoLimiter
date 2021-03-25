package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	limiter "lmf.mortal.com/GoLimiter"
	"lmf.mortal.com/GoLogs"

	"github.com/bitly/go-simplejson"
)

// 创建一个配置结构体，包含所有的配置对象
type Config struct {
	logs.LogConfig        `json:"log_config"`     // 日志服务配置
	limiter.LimiterConfig `json:"limiter_config"` // 限流服务配置
}

var (
	ConfigInstance *Config          // ConfigInstance 当前环境配置信息
	ConfigJson     *simplejson.Json // origin File 将配置信息解成 json
)

// InitConfig 初始化配置文件
// 日志服务此时尚未启动，所有的启动错误通过 panic 汇报
// 配置初始化失败，则系统无法启动，直接 panic
func InitConfig(file string) {
	confContent, err := ioutil.ReadFile(file) // 读取文件信息
	if err != nil {
		panic(fmt.Sprintf("[Logs] create new config error: %+v\n", err))
	}

	var conf Config
	err = json.Unmarshal(confContent, &conf) // 解析到 Config 结构体
	if err != nil {
		panic(fmt.Sprintf("[Logs] json unmarshal error: %+v\n", err))
	}
	ConfigInstance = &conf // 赋值到 ConfigInstance

	confJson, err := simplejson.NewJson(confContent) // 解析到 Json
	if err != nil {
		panic(fmt.Sprintf("[Logs] json unmarshal error: %+v\n", err))
	}
	ConfigJson = confJson // 赋值到 ConfigJson
}
