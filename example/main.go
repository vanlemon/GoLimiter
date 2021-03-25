package main

import (
	"fmt"
	"math/rand"
	"time"

	"lmf.mortal.com/GoLimiter/example/config"
	"lmf.mortal.com/GoLogs"
	"lmf.mortal.com/GoLogs/util"

	limiter "lmf.mortal.com/GoLimiter"
)

// 初始化配置，启动日志服务
func InitConfig() {
	// 加载配置文件
	config.InitConfig(util.GetExecPath() + "/../conf/limiter_example_dev.json")
	// 初始化日志服务
	logs.InitDefaultLogger(config.ConfigInstance.LogConfig)
	// 初始化限流器中间件
	limiter.InitOverLoadMiddleWare(
		config.ConfigInstance.LimiterConfig,
		config.ConfigJson.Get("limiter_config"),
		nil)
}

func main() {
	InitConfig()
	for i := 0; i < 100; i++ {
		go func() {
			time.Sleep(time.Millisecond * time.Duration(rand.Int63n(1000*10)))
			fmt.Println("Method1:", limiter.CanPass(logs.TestCtx(), "Method1"))
		}()
		go func() {
			time.Sleep(time.Millisecond * time.Duration(rand.Int63n(1000*10)))
			fmt.Println("Method2:", limiter.CanPass(logs.TestCtx(), "Method2"))
		}()
		go func() {
			time.Sleep(time.Millisecond * time.Duration(rand.Int63n(1000*10)))
			fmt.Println("Method3:", limiter.CanPass(logs.TestCtx(), "Method3"))
		}()
	}
	time.Sleep(time.Second * 10)
}
