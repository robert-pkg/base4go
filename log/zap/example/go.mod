module example

go 1.24.0

replace github.com/robert-pkg/base4go => ../../../

require (
	github.com/robert-pkg/base4go v0.0.0
	go.uber.org/zap v1.27.1
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require go.uber.org/multierr v1.10.0 // indirect
