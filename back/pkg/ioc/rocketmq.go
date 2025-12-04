package ioc

import (
	"context"
	"fmt"

	rocketmq "github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

// InitRocketMQProducer 根据配置初始化一个 RocketMQ Producer
// group 为生产者分组名称，通常按业务划分，例如 "order-producer"
func InitRocketMQProducer(cfg RocketMQConfig, group string) (rocketmq.Producer, error) {
	if cfg == nil {
		return nil, fmt.Errorf("rocketmq config is nil")
	}

	nameServers := cfg.GetNameServers()
	if len(nameServers) == 0 {
		return nil, fmt.Errorf("rocketmq nameservers is empty")
	}
	if group == "" {
		return nil, fmt.Errorf("rocketmq group is empty")
	}

	opts := []producer.Option{
		producer.WithNameServer(nameServers),
		producer.WithGroupName(group),
	}

	// 可选的 namespace
	if ns := cfg.GetNamespace(); ns != "" {
		opts = append(opts, producer.WithNamespace(ns))
	}

	// 可选的访问凭证
	if ak := cfg.GetAccessKey(); ak != "" {
		opts = append(opts, producer.WithCredentials(primitive.Credentials{
			AccessKey: ak,
			SecretKey: cfg.GetSecretKey(),
		}))
	}

	p, err := rocketmq.NewProducer(opts...)
	if err != nil {
		return nil, fmt.Errorf("new rocketmq producer: %w", err)
	}

	if err := p.Start(); err != nil {
		return nil, fmt.Errorf("start rocketmq producer: %w", err)
	}

	return p, nil
}

// ShutdownRocketMQProducer 优雅关闭 Producer
func ShutdownRocketMQProducer(p rocketmq.Producer) {
	if p == nil {
		return
	}
	_ = p.Shutdown()
}

// InitRocketMQConsumer 根据配置初始化一个 RocketMQ PushConsumer
// group 为消费者分组名称；topics 为需要订阅的主题列表
// handler 为业务处理函数，返回 error 时 RocketMQ 会按配置进行重试
func InitRocketMQConsumer(cfg RocketMQConfig, group string, topics []string, handler func(context.Context, *primitive.MessageExt) error) (rocketmq.PushConsumer, error) {
	if cfg == nil {
		return nil, fmt.Errorf("rocketmq config is nil")
	}
	if len(topics) == 0 {
		return nil, fmt.Errorf("rocketmq topics is empty")
	}
	if group == "" {
		return nil, fmt.Errorf("rocketmq group is empty")
	}

	nameServers := cfg.GetNameServers()
	if len(nameServers) == 0 {
		return nil, fmt.Errorf("rocketmq nameservers is empty")
	}

	opts := []consumer.Option{
		consumer.WithNameServer(nameServers),
		consumer.WithGroupName(group),
		consumer.WithConsumeFromWhere(consumer.ConsumeFromLastOffset),
	}

	// 可选的 namespace
	if ns := cfg.GetNamespace(); ns != "" {
		opts = append(opts, consumer.WithNamespace(ns))
	}

	// 可选的访问凭证
	if ak := cfg.GetAccessKey(); ak != "" {
		opts = append(opts, consumer.WithCredentials(primitive.Credentials{
			AccessKey: ak,
			SecretKey: cfg.GetSecretKey(),
		}))
	}

	c, err := rocketmq.NewPushConsumer(opts...)
	if err != nil {
		return nil, fmt.Errorf("new rocketmq consumer: %w", err)
	}

	// 订阅多个 topic，使用全量 tag（*），后续如果需要再细化 tag 区分
	for _, topic := range topics {
		if topic == "" {
			continue
		}
		if err := c.Subscribe(topic, consumer.MessageSelector{}, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
			for _, msg := range msgs {
				if handler != nil {
					if err := handler(ctx, msg); err != nil {
						// 返回 ConsumeRetryLater 让 RocketMQ 进行重试
						return consumer.ConsumeRetryLater, err
					}
				}
			}
			return consumer.ConsumeSuccess, nil
		}); err != nil {
			_ = c.Shutdown()
			return nil, fmt.Errorf("subscribe topic %s: %w", topic, err)
		}
	}

	if err := c.Start(); err != nil {
		_ = c.Shutdown()
		return nil, fmt.Errorf("start rocketmq consumer: %w", err)
	}

	return c, nil
}

// ShutdownRocketMQConsumer 优雅关闭 Consumer
func ShutdownRocketMQConsumer(c rocketmq.PushConsumer) {
	if c == nil {
		return
	}
	_ = c.Shutdown()
}
