# GoLimiter

使用示例：[example](https://github.com/lilinxi/GoLimiter/blob/master/example/main.go)

敏感数据已用`something`脱敏。

---

## 功能说明

基于方法名配置的限流，支持分布式全局限流和本地局部限流。

1. 分布式全局限流
    - 简单 redis 限流器
    - 步进 redis 限流器
2. 本地局部限流
    - 漏桶限流器
    - 令牌桶限流器

---

## 接口说明

```go
/**
添加限流器

- limiterConfig：限流器全局配置
- configJson：配置 json 字符串，应为局部配置信息而非全局配置信息，应传入 configJson.Get("limiter_config")
- redisClient：redis 客户端
*/
func InitOverLoadMiddleWare(limiterConfig LimiterConfig, configJson *simplejson.Json, redisClient *redis.Client) {
	
/**
请求是否通过限流器

- method：方法名
- local：是否使用局部限流
*/
func CanPass(ctx context.Context, method string) bool {
```

---

## 限流器 Gin 中间件实例

```go
// 限流器中间件
func LimiterMiddleware() gin.HandlerFunc {
	return func(gctx *gin.Context) {
		// step1: 获取请求 ctx
		ctxValue, exists := gctx.Get(cconst.CtxKey)
		ctx, ok := ctxValue.(context.Context)
		if !exists || !ok {
			panic("ctx not exists")
		}

		// step2: 打印请求日志
		methodName := gctx.Request.Method + ":" + gctx.Request.URL.String()
		pass := limiter.CanPass(ctx, methodName)
		if pass {
			logs.CtxInfo(ctx, "[Access] %s CanPass? \t %+v", methodName, pass)
			// step3: 执行服务
			gctx.Next()
		} else {
			logs.CtxWarn(ctx, "[Access] %s CanPass? \t %+v", methodName, pass)
			// step3: 服务器超载，直接返回
			gctx.JSON(http.StatusOK, util.Overload())
			gctx.Abort()
		}
	}
}
```

---

## 限流器配置

```go
"limiter_config": {
    "service_name": "limiter.example.dev",  // 服务名
    "default_pass": true,                   // 未设置限额默认通过还是默认拦截
    "global_on": false,                     // 是否开启分布式全局限流，还是启用本地局部限流
    "global_limiter_type": "simple",        // 分布式全局限流类型，simple：简单redis限流器；cache：步进redis限流器
    "local_limiter_type": "token",          // 本地局部限流类型：leaky：漏桶限流器；token：令牌桶限流器
    "method_qps_limit": {                   // 接口限流配置
      "Method1": {          // 接口名
        "step": 0,          // 限流步长
        "global_qps": 0,    // 分布式限流限额
        "local_qps": 2      // 本地限流限额
      },
      "Method2": {
        "step": 0,
        "global_qps": 0,
        "local_qps": 1
      }
    }
  },
```
