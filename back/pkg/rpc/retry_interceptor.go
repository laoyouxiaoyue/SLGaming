package rpc

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryRetryInterceptor 创建一元RPC调用的重试拦截器
//
// 这是一个 gRPC 客户端拦截器，会自动对失败的请求进行重试。
// 使用指数退避策略，避免对服务端造成过大压力。
//
// 工作流程:
//  1. 发送请求
//  2. 如果成功，直接返回
//  3. 如果失败且错误码可重试，等待退避时间后重试
//  4. 重复直到成功或达到最大重试次数
//
// 使用示例:
//
//	client := zrpc.MustNewClient(zrpc.RpcClientConf{...},
//	    zrpc.WithUnaryClientInterceptor(UnaryRetryInterceptor(DefaultRetryOptions())),
//	)
func UnaryRetryInterceptor(opts RetryOptions) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, callOpts ...grpc.CallOption) error {

		// 如果禁用重试，直接调用
		if opts.MaxRetries <= 0 {
			return invoker(ctx, method, req, reply, cc, callOpts...)
		}

		var lastErr error
		var attempt int

		// 创建带总超时的上下文，防止重试时间过长
		retryCtx, cancel := context.WithTimeout(ctx, opts.RetryTimeout)
		defer cancel()

		// 重试循环: attempt=0 是首次请求, attempt=1,2,3... 是重试
		for attempt = 0; attempt <= opts.MaxRetries; attempt++ {
			// 非首次请求需要等待退避时间
			if attempt > 0 {
				backoff := opts.CalculateBackoff(attempt - 1)
				select {
				case <-retryCtx.Done():
					// 总超时已到，返回最后的错误
					if lastErr != nil {
						return lastErr
					}
					return retryCtx.Err()
				case <-time.After(backoff):
					// 等待退避时间后继续
				}
				logx.WithContext(ctx).Debugf("[rpc_retry] retrying request: method=%s, attempt=%d, backoff=%v", method, attempt, backoff)
			}

			// 执行实际的RPC调用
			err := invoker(retryCtx, method, req, reply, cc, callOpts...)
			if err == nil {
				// 成功
				if attempt > 0 {
					logx.WithContext(ctx).Infof("[rpc_retry] request succeeded after retry: method=%s, attempts=%d", method, attempt+1)
				}
				return nil
			}

			lastErr = err

			// 从错误中提取gRPC状态码
			st, ok := status.FromError(err)
			if !ok {
				// 非gRPC错误，直接返回
				return err
			}

			// 检查错误码是否可重试
			if !opts.IsRetryable(st.Code()) {
				logx.WithContext(ctx).Debugf("[rpc_retry] non-retryable error: method=%s, code=%s, error=%v", method, st.Code(), err)
				return err
			}

			// 记录可重试错误，继续下一次重试
			logx.WithContext(ctx).Infof("[rpc_retry] retryable error: method=%s, attempt=%d, code=%s, error=%v", method, attempt+1, st.Code(), err)
		}

		// 达到最大重试次数仍未成功
		logx.WithContext(ctx).Errorf("[rpc_retry] max retries exceeded: method=%s, attempts=%d, last_error=%v", method, attempt, lastErr)
		return lastErr
	}
}

// StreamRetryInterceptor 创建流式RPC调用的重试拦截器
//
// 注意：流式RPC的重试只在建立连接阶段有效。
// 一旦流建立成功，后续的数据传输不会自动重试。
//
// 适用于：
//   - 服务端流
//   - 客户端流
//   - 双向流
//
// 使用示例:
//
//	client := zrpc.MustNewClient(zrpc.RpcClientConf{...},
//	    zrpc.WithStreamClientInterceptor(StreamRetryInterceptor(DefaultRetryOptions())),
//	)
func StreamRetryInterceptor(opts RetryOptions) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, callOpts ...grpc.CallOption) (grpc.ClientStream, error) {

		if opts.MaxRetries <= 0 {
			return streamer(ctx, desc, cc, method, callOpts...)
		}

		var lastErr error
		var attempt int

		retryCtx, cancel := context.WithTimeout(ctx, opts.RetryTimeout)
		defer cancel()

		for attempt = 0; attempt <= opts.MaxRetries; attempt++ {
			if attempt > 0 {
				backoff := opts.CalculateBackoff(attempt - 1)
				select {
				case <-retryCtx.Done():
					if lastErr != nil {
						return nil, lastErr
					}
					return nil, retryCtx.Err()
				case <-time.After(backoff):
				}
				logx.WithContext(ctx).Debugf("[rpc_retry] retrying stream: method=%s, attempt=%d, backoff=%v", method, attempt, backoff)
			}

			// 尝试建立流连接
			stream, err := streamer(retryCtx, desc, cc, method, callOpts...)
			if err == nil {
				if attempt > 0 {
					logx.WithContext(ctx).Infof("[rpc_retry] stream succeeded after retry: method=%s, attempts=%d", method, attempt+1)
				}
				return stream, nil
			}

			lastErr = err

			st, ok := status.FromError(err)
			if !ok {
				return nil, err
			}

			if !opts.IsRetryable(st.Code()) {
				return nil, err
			}

			logx.WithContext(ctx).Infof("[rpc_retry] stream retryable error: method=%s, attempt=%d, code=%s, error=%v", method, attempt+1, st.Code(), err)
		}

		return nil, lastErr
	}
}

// wrappedClientStream 包装gRPC客户端流
// 用于未来可能的功能扩展（如流级别的重试）
type wrappedClientStream struct {
	grpc.ClientStream
}

func (w *wrappedClientStream) SendMsg(m interface{}) error {
	return w.ClientStream.SendMsg(m)
}

func (w *wrappedClientStream) RecvMsg(m interface{}) error {
	return w.ClientStream.RecvMsg(m)
}

// IsRetryableError 检查错误是否可重试
//
// 这是一个辅助函数，用于业务代码判断是否需要手动重试。
//
// 示例:
//
//	err := client.Call(ctx, req)
//	if rpc.IsRetryableError(err, opts) {
//	    // 可以考虑手动重试
//	}
func IsRetryableError(err error, opts RetryOptions) bool {
	if err == nil {
		return false
	}

	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	return opts.IsRetryable(st.Code())
}

// IsUnavailable 检查错误是否为 Unavailable
// 通常表示服务不可用、网络问题或服务重启中
func IsUnavailable(err error) bool {
	return isCode(err, codes.Unavailable)
}

// IsDeadlineExceeded 检查错误是否为 DeadlineExceeded
// 通常表示请求超时
func IsDeadlineExceeded(err error) bool {
	return isCode(err, codes.DeadlineExceeded)
}

// IsAborted 检查错误是否为 Aborted
// 通常表示事务中止或并发冲突
func IsAborted(err error) bool {
	return isCode(err, codes.Aborted)
}

// IsResourceExhausted 检查错误是否为 ResourceExhausted
// 通常表示服务限流或资源耗尽
func IsResourceExhausted(err error) bool {
	return isCode(err, codes.ResourceExhausted)
}

// isCode 检查错误是否为指定的gRPC错误码
func isCode(err error, code codes.Code) bool {
	if err == nil {
		return false
	}
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	return st.Code() == code
}
