package consts

import "google.golang.org/grpc/status"

var (
	ErrPasswordNotEqual = status.Error(20001, "密码错误")
	ErrCodeNotFound     = status.Error(20002, "验证码已过期")
	ErrCodeNotEqual     = status.Error(20003, "验证码错误")
	ErrHaveExist        = status.Error(20004, "邮箱已被注册")
	ErrNotFound         = status.Error(20006, "数据不存在")
	ErrInvalidObjectId  = status.Error(20007, "ID格式错误")
)
