package limiter

import (
	logs "lmf.mortal.com/GoLogs"
)

// 限流器接口
type LimiterImpl interface {
	// 限流器限额，限制的每秒请求数
	SetLimit(newLimit int64) (oldLimit int64) // 应为原子操作
	GetLimit() (oldLimit int64)

	// 限流器步长，默认不使用步长，步长为0，当集群限额很大时，使用分布式限流器可设置步长减缓 redis 压力
	SetStep(newLimit int64) (oldStep int64) // 应为原子操作
	GetStep() (oldStep int64)

	// 在同一个原子操作内同时设置限额和步长
	SetLimitAndStep(newLimit int64, newStep int64) (oldLimit int64, oldStep int64)

	// 当前请求是否可通过限流器
	CanPass() (pass bool)
}

// 限流器默认实现
type Limiter struct {
}

func (s *Limiter) SetLimit(newLimit int64) (oldLimit int64) {
	logs.CtxWarn(logs.SysCtx, "[Running Limiter] SetLimit no impl")
	return
}

func (s *Limiter) GetLimit() (oldLimit int64) {
	logs.CtxWarn(logs.SysCtx, "[Running Limiter] GetLimit no impl")
	return
}

func (s *Limiter) SetStep(new int64) (oldStep int64) {
	logs.CtxWarn(logs.SysCtx, "[Running Limiter] SetStep no impl")
	return
}

func (s *Limiter) GetStep() (oldStep int64) {
	logs.CtxWarn(logs.SysCtx, "[Running Limiter] GetStep no impl")
	return
}

func (s *Limiter) SetLimitAndStep(newLimit int64, newStep int64) (oldLimit int64, oldStep int64) {
	logs.CtxWarn(logs.SysCtx, "[Running Limiter] SetLimitAndStep no impl")
	return
}

func (s *Limiter) CanPass() (pass bool) {
	logs.CtxFatal(logs.SysCtx, "[Running Limiter] CanPass no impl")
	panic("[Running Limiter] CanPass no impl")
}
