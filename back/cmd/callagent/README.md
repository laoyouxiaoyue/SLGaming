# Agent 服务测试客户端

用于测试 Agent 服务的 RPC 接口。

## 使用方法

### 1. 启动 Agent 服务

首先需要启动 Agent 服务：

```bash
cd E:\workspace\go\SLGaming\back\services\agent
go run agent.go -f etc/agent.yaml
```

或者：

```bash
cd E:\workspace\go\SLGaming\back\services\agent
.\agent.exe -f etc/agent.yaml
```

### 2. 运行测试客户端

#### 添加陪玩信息到向量数据库

```bash
cd E:\workspace\go\SLGaming\back\cmd\callagent
.\callagent.exe -action=add -user_id=2001 -gender=male -age=22 -game="wangzhe" -desc="test description" -price=50 -rating=4.8
```

参数说明：
- `-endpoint`: Agent RPC 服务地址（默认: 127.0.0.1:8080）
- `-action`: 操作类型（add|recommend）
- `-user_id`: 用户ID
- `-gender`: 性别（male/female/other）
- `-age`: 年龄
- `-game`: 游戏技能
- `-desc`: 描述文本
- `-price`: 每小时价格
- `-rating`: 评分（0-5）

#### 推荐陪玩

```bash
.\callagent.exe -action=recommend -text="我想要一个王者荣耀的陪玩" -user_id=1001
```

参数说明：
- `-text`: 用户输入文本
- `-user_id`: 用户ID（可选）

## 示例

### 示例1：添加陪玩信息

```bash
.\callagent.exe -action=add \
  -user_id=2001 \
  -gender=male \
  -age=22 \
  -game="王者荣耀" \
  -desc="专业王者荣耀陪玩，擅长打野和ADC，段位王者50星" \
  -price=50 \
  -rating=4.8
```

### 示例2：推荐陪玩

```bash
.\callagent.exe -action=recommend \
  -text="我想要一个王者荣耀的陪玩" \
  -user_id=1001
```

## 注意事项

1. 确保 Agent 服务已启动并监听在指定端口
2. 确保 Milvus 服务已启动（默认: 127.0.0.1:19530）
3. 确保 LLM 配置正确（APIKey 和 Model）
4. 首次运行会自动创建 Collection（如果不存在）
