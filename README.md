# 21-dhcp-ipam-control

高可用DHCP与IP地址管理服务，提供IPv4/IPv6网络空间、子网和地址池管理，租约分配与释放、冲突记录、配置发布以及双节点事件复制模拟。默认使用内存存储以便本地运行，生产部署可替换为PostgreSQL适配器。

## 启动

```sh
go run ./cmd/server
```

环境变量：`HTTP_ADDR`（默认`:8080`）、`DHCPV4_ADDR`（`:6767`）、`DHCPV6_ADDR`（`:6768`）、`NODE_ID`、`HA_MODE`、`LEASE_TTL`、`AUTH_TOKEN`。

## API示例

```sh
curl http://localhost:8080/healthz
curl -X POST http://localhost:8080/v1/networks -H 'content-type: application/json' -d '{"id":"net-1","name":"lab","cidr":"10.10.0.0/16"}'
curl -X POST http://localhost:8080/v1/subnets -H 'content-type: application/json' -d '{"id":"sub-1","networkId":"net-1","cidr":"10.10.1.0/24"}'
curl -X POST http://localhost:8080/v1/pools -H 'content-type: application/json' -d '{"id":"pool-1","subnetId":"sub-1","start":"10.10.1.10","end":"10.10.1.200"}'
curl -X POST http://localhost:8080/v1/leases/allocate -H 'content-type: application/json' -d '{"id":"lease-1","poolId":"pool-1","clientId":"00:11:22:33:44:55","family":"ipv4","address":"10.10.1.20"}'
curl -X POST http://localhost:8080/v1/leases/lease-1/release
```

`GET /v1/leases?cursor=&limit=20`使用游标分页；`/metrics`暴露Prometheus文本指标。DHCPv4 UDP端口实现报文解析和DISCOVER的OFFER模拟，DHCPv6支持SOLICIT的ADVERTISE模拟。服务收到SIGTERM后停止后台worker和UDP监听并优雅关闭HTTP服务器。

## 目录

`cmd`启动入口；`internal/*`按领域及domain/application/adapter/infrastructure分层；`api`包含OpenAPI和gRPC契约；`migrations`为PostgreSQL表结构；`deploy`含Compose；`scripts`含启动和迁移脚本。
