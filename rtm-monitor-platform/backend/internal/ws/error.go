package ws

import "errors"

var (
	ErrHostProxyNotConnected  = errors.New("hostProxy未连接")
	ErrHostProxySendBufferFull = errors.New("hostProxy发送缓冲区已满")
	ErrSessionNotFound         = errors.New("会话不存在")
	ErrSessionSendBufferFull   = errors.New("会话发送缓冲区已满")
	ErrCallbackTimeout         = errors.New("回调超时")
)
