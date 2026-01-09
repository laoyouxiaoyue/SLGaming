# 创建 RocketMQ Topic - order_events

## 快速解决方案

根据你的配置（NameServer: `120.26.29.194:9876`），使用以下命令创建 Topic：

### 方法1：使用 RocketMQ 控制台（推荐）

1. 访问 RocketMQ 控制台：`http://120.26.29.194:8080`（如果部署了控制台）
2. 登录后进入 **Topic** 管理
3. 点击 **创建 Topic**
4. 填写信息：
   - **Topic**: `order_events`
   - **读写队列数**: 4
   - **权限**: 6 (读写)
5. 点击确认

### 方法2：使用命令行工具

#### Windows:

```powershell
# 进入 RocketMQ bin 目录（根据实际安装路径调整）
cd C:\rocketmq\bin

# 创建 Topic
.\mqadmin.cmd updateTopic -n 120.26.29.194:9876 -c DefaultCluster -t order_events -w 4 -r 4
```

#### Linux/Mac:

```bash
# 进入 RocketMQ bin 目录
cd /opt/rocketmq/bin

# 创建 Topic
./mqadmin updateTopic -n 120.26.29.194:9876 -c DefaultCluster -t order_events -w 4 -r 4
```

### 方法3：使用提供的脚本

#### Windows:
```powershell
cd back/services/order/scripts
# 设置环境变量（可选，不设置会使用默认值）
$env:ROCKETMQ_NAMESERVER="120.26.29.194:9876"
$env:ROCKETMQ_HOME="C:\rocketmq"  # 根据实际路径修改
.\create_topic.bat
```

#### Linux/Mac:
```bash
cd back/services/order/scripts
chmod +x create_topic.sh
# 设置环境变量（可选）
export ROCKETMQ_NAMESERVER="120.26.29.194:9876"
export ROCKETMQ_HOME="/opt/rocketmq"  # 根据实际路径修改
./create_topic.sh
```

## 验证 Topic 是否创建成功

```bash
# 查看所有 Topic
mqadmin topicList -n 120.26.29.194:9876

# 查看特定 Topic 信息
mqadmin topicStatus -n 120.26.29.194:9876 -t order_events
```

## 如果创建失败

1. **确认 RocketMQ 服务已启动**
   - NameServer 是否运行：检查 `120.26.29.194:9876` 是否可访问
   - Broker 是否运行：检查 Broker 进程

2. **检查网络连接**
   ```bash
   # 测试 NameServer 连接
   telnet 120.26.29.194 9876
   # 或
   nc -zv 120.26.29.194 9876
   ```

3. **检查集群名称**
   - 如果集群名称不是 `DefaultCluster`，需要修改命令中的 `-c` 参数

4. **检查权限**
   - 如果启用了 ACL，需要提供 AccessKey 和 SecretKey

## 重启服务

创建 Topic 后，重启订单服务即可：

```bash
# 停止当前服务（Ctrl+C）
# 然后重新启动
go run back/services/order/order.go -f back/services/order/etc/order.yaml
```

## 说明

- **Topic 名称**: `order_events`（固定，代码中定义）
- **队列数**: 4（可根据实际负载调整）
- **集群名称**: `DefaultCluster`（默认值，根据实际情况调整）

创建成功后，Consumer 启动错误会消失，服务可以正常接收订单事件。












