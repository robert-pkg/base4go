# gateway

```
make build   # 编译
make run     # 本地运行
make test    # 运行测试
make fmt     # 格式化
make docker  # 打镜像
```

创建默认的日志目录

```
sudo mkdir -p /var/log/gateway
chown -R robert:robert ./gateway
```

手工运行：

```
# 环境变量覆盖配置文件中的信息
GATEWAY1_SERVER_PORT=9095 ./bin/gateway -config=./gateway.yaml -env_prerix=gateway1 -output_console=true

./bin/gateway -config=./gateway.yaml -output_console=true
```
