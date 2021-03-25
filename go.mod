module lmf.mortal.com/GoLimiter

go 1.15

require (
	github.com/bitly/go-simplejson v0.5.0
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/juju/ratelimit v1.0.1
	github.com/onsi/ginkgo v1.15.0 // indirect
	github.com/onsi/gomega v1.10.5 // indirect
	github.com/sirupsen/logrus v1.7.1
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/ratelimit v0.1.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	lmf.mortal.com/GoLogs v0.0.0-00010101000000-000000000000
)

replace lmf.mortal.com/GoLogs => ../GoLogs
