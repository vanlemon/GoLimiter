package main

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"lmf.mortal.com/GoLimiter/example/config"

	limiter "lmf.mortal.com/GoLimiter"
	logs "lmf.mortal.com/GoLogs"
)

func init() {
	// 加载配置文件
	config.InitConfig("./conf/limiter_example_dev.json")
	// 初始化日志服务
	logs.InitDefaultLogger(config.ConfigInstance.LogConfig)
	// 初始化限流器中间件
	limiter.InitOverLoadMiddleWare(
		config.ConfigInstance.LimiterConfig,
		config.ConfigJson.Get("limiter_config"),
		nil)
}

func TestMethod1(t *testing.T) {
	for i := 0; i < 100; i++ {
		go func() {
			time.Sleep(time.Millisecond * time.Duration(rand.Int63n(1000*10)))
			fmt.Println("Method1:", limiter.CanPass(logs.TestCtx(), "Method1"))
		}()
	}
	time.Sleep(time.Second * 10)
}

func TestMethod2(t *testing.T) {
	for i := 0; i < 100; i++ {
		go func() {
			time.Sleep(time.Millisecond * time.Duration(rand.Int63n(1000*10)))
			fmt.Println("Method2:", limiter.CanPass(logs.TestCtx(), "Method2"))
		}()
	}
	time.Sleep(time.Second * 10)
}
