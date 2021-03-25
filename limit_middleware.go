package limiter

import (
	"context"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/go-redis/redis"

	logs "lmf.mortal.com/GoLogs"
)

const (
	GlobalSimpleRedisLimiter = "simple" // 全局简单 redis 限流器
	GlobalCacheRedisLimiter  = "cache"  // 全局步进 redis 限流器
	LocalLeakyLimiter        = "leaky"  // 本地漏桶限流器
	LocalTokenLimiter        = "token"  // 本地令牌桶限流器
)

type LimiterConfig struct { // 限流器配置
	ServiceName       string `json:"service_name"`        // 服务名
	DefaultPass       bool   `json:"default_pass"`        // 未识别方法默认通过
	GlobalOn          bool   `json:"global_on"`           // 是否使用分布式全局限流
	GlobalLimiterType string `json:"global_limiter_type"` // 分布式全局限流器类型
	LocalLimiterType  string `json:"local_limiter_type"`  // 本地限流器类型
}

// 限流信息
type OverloadInfo struct {
	Step int64 // 步长

	GlobalQps int64 // 全局限额 qps
	LocalQps  int64 // 局部限额 qps

	GlobalLimiter LimiterImpl // 全局限流器
	LocalLimiter  LimiterImpl //局部限流器
}

var globalLimiterConfig LimiterConfig              // 限流器全局配置
var globalOverloadInfoMap map[string]*OverloadInfo // 方法名和限流信息的关联 map

// step1: 校验入参
func checkAddOverLoadMiddleWareParam(limiterConfig LimiterConfig, configJson *simplejson.Json, redisClient *redis.Client) bool {
	// 配置文件为空
	if configJson == nil {
		logs.CtxError(logs.SysCtx, "[Bootstrap Limiter] checkAddOverLoadMiddleWareParam configJson not init")
		return false
	}
	// 若全局限流器开启，则应存在 redis 客户端
	if limiterConfig.GlobalOn && redisClient == nil {
		logs.CtxError(logs.SysCtx, "[Bootstrap Limiter] checkAddOverLoadMiddleWareParam redisClient not init")
		return false
	}
	return true
}

// step2.1: 从 configJson 中获取配置信息，将配置信息解析成`方法名 -> 限流配置 map`的形式
func loadConfigSetting(configJson *simplejson.Json) map[string]map[string]int64 {
	// 创建：方法名 -> 限流配置 map
	methodQpsLimitMap := map[string]map[string]int64{}
	// 获取原始配置信息
	methodQpsLimit := configJson.Get("method_qps_limit")
	// 遍历每一个限流配置中的方法名
	for method, _ := range methodQpsLimit.MustMap() {
		qpsLimitMap := make(map[string]int64)        // 创建限流配置 map
		qpsLimitJson := methodQpsLimit.Get(method)   // 获取每个方法的原始配置信息
		for key, _ := range qpsLimitJson.MustMap() { // 遍历每一个配置 key
			qpsLimitMap[key] = qpsLimitJson.Get(key).MustInt64() // 写入每一个 key 对应的 int64 值
		}
		methodQpsLimitMap[method] = qpsLimitMap // 关联每个方法名和配置信息 map
	}
	logs.CtxInfo(logs.SysCtx, "[Bootstrap Limiter] method-qpsLimit map: %+v", methodQpsLimitMap)
	return methodQpsLimitMap
}

// step2: 从 configJson 中获取配置信息，将配置信息解析成`方法名 -> 限流配置结构体`的形式
func InitQpsAndStepFromConfig(configJson *simplejson.Json) {
	// 将配置信息解析成`方法名 -> 限流配置 map`的形式
	methodQpsLimitMap := loadConfigSetting(configJson)
	// 创建：方法名 -> 限流配置结构体，这里使用的是全局变量
	globalOverloadInfoMap = make(map[string]*OverloadInfo)
	for method, qpsLimitMap := range methodQpsLimitMap {
		// 解析限流配置
		step := qpsLimitMap["step"]
		globalQps := qpsLimitMap["global_qps"]
		localQps := qpsLimitMap["local_qps"]
		if step < 0 || globalQps < 0 || localQps < 0 {
			logs.CtxFatal(logs.SysCtx, "[Bootstrap Limiter] InitQpsAndStepFromConfig bad step: %d, globalQps: %d, localQps: %d for method: %s", step, globalQps, localQps, method)
			panic(fmt.Sprintf("[Bootstrap Limiter] InitQpsAndStepFromConfig bad step: %d, globalQps: %d, localQps: %d for method: %s"))
		}
		// 创建限流配置结构体
		overloadInfo := &OverloadInfo{
			Step:      step,
			GlobalQps: globalQps,
			LocalQps:  localQps,
		}
		// 将限流配置结构体关联到方法名
		globalOverloadInfoMap[method] = overloadInfo
	}
	logs.CtxInfo(logs.SysCtx, "[Bootstrap Limiter] global method-overloadInfo map: %+v", globalOverloadInfoMap)
}

// step3: 初始化限流器
func InitLimiter(limiterConfig LimiterConfig, redisClient *redis.Client) {
	// 遍历方法名和配置结构体
	for method, overloadInfo := range globalOverloadInfoMap {
		// 方法唯一键名：limiter_name + service_name + method_name
		methodKey := limiterConfig.ServiceName + "." + method
		// 初始化分布式全局限流器
		if limiterConfig.GlobalOn {
			switch limiterConfig.GlobalLimiterType {
			case GlobalSimpleRedisLimiter:
				overloadInfo.GlobalLimiter = NewSimpleRedisLimiter(methodKey, overloadInfo.GlobalQps, redisClient)
				break
			case GlobalCacheRedisLimiter:
				// TODO
				logs.CtxFatal(logs.SysCtx, "[Bootstrap Limiter] GlobalCacheRedisLimiter no impl")
				panic("[Bootstrap Limiter] GlobalCacheRedisLimiter no impl")
			default:
				logs.CtxFatal(logs.SysCtx, "[Bootstrap Limiter] GlobalLimiterType error: %+v", limiterConfig.GlobalLimiterType)
				panic(fmt.Sprintf("[Bootstrap Limiter] GlobalLimiterType error: %+v", limiterConfig.GlobalLimiterType))
			}
		}
		// 初始化本地限流器
		switch limiterConfig.LocalLimiterType {
		case LocalLeakyLimiter:
			overloadInfo.LocalLimiter = NewLeakyLimiter(methodKey, overloadInfo.LocalQps)
			break
		case LocalTokenLimiter:
			overloadInfo.LocalLimiter = NewTokenLimiter(methodKey, overloadInfo.LocalQps)
		default:
			logs.CtxFatal(logs.SysCtx, "[Bootstrap Limiter] LocalLimiterType error: %+v", limiterConfig.LocalLimiterType)
			panic(fmt.Sprintf("[Bootstrap Limiter] LocalLimiterType error: %+v", limiterConfig.LocalLimiterType))
		}
	}
}

/**
添加限流器

- limiterConfig：限流器全局配置
- configJson：配置 json 字符串，应为局部配置信息而非全局配置信息，应传入 configJson.Get("limiter_config")
- redisClient：redis 客户端
*/
func InitOverLoadMiddleWare(limiterConfig LimiterConfig, configJson *simplejson.Json, redisClient *redis.Client) {
	logs.CtxInfo(logs.SysCtx, "[Bootstrap Limiter] InitOverLoadMiddleWare limiterConfig: %+v", limiterConfig)
	// step1: 加载限流器全局配置
	globalLimiterConfig = limiterConfig
	// step2: 校验入参
	if !checkAddOverLoadMiddleWareParam(limiterConfig, configJson, redisClient) {
		logs.CtxFatal(logs.SysCtx, "[Bootstrap Limiter] InitOverLoadMiddleWare checkAddOverLoadMiddleWareParam failed")
		panic("[Bootstrap Limiter] InitOverLoadMiddleWare checkAddOverLoadMiddleWareParam failed")
	}
	// step3: 从 configJson 中获取配置信息，解析限流配置结构体
	InitQpsAndStepFromConfig(configJson)
	// step4: 初始化限流器
	InitLimiter(limiterConfig, redisClient)
}

/**
请求是否通过限流器

- method：方法名
- local：是否使用局部限流
*/
func CanPass(ctx context.Context, method string) bool {
	overload, ok := globalOverloadInfoMap[method] // 根据方法名获取限流信息

	// 限流信息不存在，根据配置默认通过或拦截
	if !ok || overload == nil {
		logs.CtxError(ctx, "[Bootstrap Limiter] method: %s, not found int limiter", method)
		return globalLimiterConfig.DefaultPass
	}

	// 开启分布式全局限流
	if globalLimiterConfig.GlobalOn {
		return overload.GlobalLimiter.CanPass()
	} else { // 使用局部本地限流
		return overload.LocalLimiter.CanPass()
	}
}
