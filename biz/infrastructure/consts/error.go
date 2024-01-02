package consts

import "google.golang.org/grpc/status"

var (
	ErrDataBase      = status.Error(10001, "数据库异常")
	ErrRedis         = status.Error(10002, "Redis异常")
	ErrEmailNotSend  = status.Error(10003, "邮件发送失败")
	ErrCaptchaCreate = status.Error(10004, "验证码生成失败")
)

var (
	ErrPasswordNotEqual = status.Error(20001, "密码错误")
	ErrCodeNotFound     = status.Error(20002, "验证码已过期")
	ErrCodeNotEqual     = status.Error(20003, "验证码错误")
	ErrHaveExist        = status.Error(20004, "邮箱已被注册")
	ErrEmailNotFound    = status.Error(20005, "邮箱不存在")
	ErrNotFound         = status.Error(20006, "数据不存在")
	ErrInvalidObjectId  = status.Error(20007, "ID格式错误")
)
