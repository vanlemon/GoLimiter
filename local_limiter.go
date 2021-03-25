package limiter

import (
	"sync"
	"time"

	logs "lmf.mortal.com/GoLogs"

	tokenlimiter "github.com/juju/ratelimit"
	leakylimiter "go.uber.org/ratelimit"
)

/**
本地限流器

- 漏桶：https://github.com/uber-go/ratelimit
- 令牌桶：https://github.com/juju/ratelimit

漏桶限流器可以会阻塞超负荷流量，使得流量更加平滑
令牌桶限流器直接丢弃超负荷流量，且可以在一定程度上应对突发流量
*/

// 本地漏桶限流器
type LeakyLimiter struct {
	Limiter
	key        string               // 限流器唯一标识
	localLimit int64                // 限流器限额
	bucket     leakylimiter.Limiter // 漏桶限流器
	mutex      sync.Mutex           // 原子信号量锁
}

func NewLeakyLimiter(key string, localLimit int64) *LeakyLimiter {
	if localLimit <= 0 { // 限额应大于零
		logs.CtxError(logs.SysCtx, "[Running Limiter] NewLeakyLimiter bad params, key: %s, localLimit: %d", key, localLimit)
		return nil
	}

	// 创建漏桶，设置每秒均匀限额；TODO int64 转 int
	thisBucket := leakylimiter.New(int(localLimit)) // per second

	return &LeakyLimiter{
		key:        "local-leaky-limiter-" + key,
		localLimit: localLimit,
		bucket:     thisBucket,
	}
}

func (s *LeakyLimiter) SetLimit(newLimit int64) (oldLimit int64) {
	// 原子操作，需要加锁
	s.mutex.Lock()
	oldLimit = s.localLimit
	s.localLimit = newLimit
	s.bucket = leakylimiter.New(int(newLimit)) // per second
	s.mutex.Unlock()

	logs.CtxInfo(logs.SysCtx, "[Running Limiter] LeakyLimiter SetLimit key: %s, old: %d, new: %d", s.key, oldLimit, newLimit)
	return oldLimit
}

func (s *LeakyLimiter) GetLimit() (oldLimit int64) {
	return s.localLimit
}

func (s *LeakyLimiter) SetLimitAndStep(newLimit int64, newStep int64) (oldLimit int64, oldStep int64) {
	return s.SetLimit(newLimit), 0
}

func (s *LeakyLimiter) CanPass() (pass bool) {
	s.bucket.Take() // 阻塞超负荷流量
	return true
}

// 本地令牌桶限流器
type TokenLimiter struct {
	Limiter
	key        string               // 限流器唯一标识
	localLimit int64                // 限流器限额
	bucket     *tokenlimiter.Bucket // 令牌桶限流器
	mutex      sync.Mutex           // 原子信号量锁
}

func NewTokenLimiter(key string, localLimit int64) *TokenLimiter {
	if localLimit <= 0 { // 限额应大于零
		logs.CtxError(logs.SysCtx, "[Running Limiter] NewTokenLimiter bad params, key: %s, localLimit: %d", key, localLimit)
		return nil
	}

	// 创建令牌桶，每个令牌被放入的时间间隔，和令牌桶的最大容量
	thisBucket := tokenlimiter.NewBucket(time.Second/time.Duration(localLimit), localLimit)

	return &TokenLimiter{
		key:        "local-token-limiter-" + key,
		localLimit: localLimit,
		bucket:     thisBucket,
	}
}

func (s *TokenLimiter) SetLimit(newLimit int64) (oldLimit int64) {
	// 原子操作，需要加锁
	s.mutex.Lock()
	oldLimit = s.localLimit
	s.localLimit = newLimit
	s.bucket = tokenlimiter.NewBucket(time.Second/time.Duration(newLimit), newLimit)
	s.mutex.Unlock()

	logs.CtxInfo(logs.SysCtx, "[Running Limiter] TokenLimiter SetLimit key: %s, old: %d, new: %d", s.key, oldLimit, newLimit)
	return oldLimit
}

func (s *TokenLimiter) GetLimit() (oldLimit int64) {
	return s.localLimit
}

func (s *TokenLimiter) SetLimitAndStep(newLimit int64, newStep int64) (oldLimit int64, oldStep int64) {
	return s.SetLimit(newLimit), 0
}

func (s *TokenLimiter) CanPass() (pass bool) {
	// 返回获取到的令牌数，每次请求获取一个令牌，返回 1 则表示获取到
	return s.bucket.TakeAvailable(1) == 1
}
