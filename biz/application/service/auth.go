package service

import (
	"context"
	"fmt"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/config"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/consts"
	usermapper "github.com/CloudStriver/cloudmind-sts/biz/infrastructure/mapper/user"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/util/email"
	gensts "github.com/CloudStriver/service-idl-gen-go/kitex_gen/cloudmind/sts"
	"github.com/google/wire"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type AuthService interface {
	CreateAuth(ctx context.Context, req *gensts.CreateAuthReq) (resp *gensts.CreateAuthResp, err error)
	CheckEmail(ctx context.Context, req *gensts.CheckEmailReq) (resp *gensts.CheckEmailResp, err error)
	SetPassword(ctx context.Context, req *gensts.SetPasswordReq) (resp *gensts.SetPasswordResp, err error)
	SendEmail(ctx context.Context, req *gensts.SendEmailReq) (resp *gensts.SendEmailResp, err error)
	Login(ctx context.Context, req *gensts.LoginReq) (resp *gensts.LoginResp, err error)
	AppendAuth(ctx context.Context, req *gensts.AppendAuthReq) (resp *gensts.AppendAuthResp, err error)
}

var AuthSet = wire.NewSet(
	wire.Struct(new(AuthServiceImpl), "*"),
	wire.Bind(new(AuthService), new(*AuthServiceImpl)),
)

type AuthServiceImpl struct {
	Config          *config.Config
	Redis           *redis.Redis
	UserMongoMapper usermapper.IUserMongoMapper
}

// 添加登录方式
func (s *AuthServiceImpl) AppendAuth(ctx context.Context, req *gensts.AppendAuthReq) (resp *gensts.AppendAuthResp, err error) {
	resp = new(gensts.AppendAuthResp)
	if err = s.UserMongoMapper.AppendAuth(ctx, req.UserId, &usermapper.Auth{
		Type:       req.AuthType,
		AppId:      req.AppId,
		UnionId:    req.UnionId,
		PlatformId: req.PlatFormId,
	}); err != nil {
		return resp, err
	}
	return resp, nil
}

// 通过某个登录方式登录
func (s *AuthServiceImpl) Login(ctx context.Context, req *gensts.LoginReq) (resp *gensts.LoginResp, err error) {
	resp = new(gensts.LoginResp)
	user, err := s.UserMongoMapper.FindOneByAuth(ctx, &usermapper.Auth{
		Type:       req.AuthType,
		AppId:      req.AppId,
		UnionId:    req.UnionId,
		PlatformId: req.PlatFormId,
	})
	if errors.Is(err, consts.ErrNotFound) {
		return resp, nil
	}
	if err != nil {
		return resp, err
	}

	if req.Password == "" {
		req.Password = consts.DefaultPassword
	}
	if user.PassWord != req.Password {
		return resp, consts.ErrPasswordNotEqual
	}

	resp.UserId = user.ID.Hex()
	return resp, nil
}

func (s *AuthServiceImpl) CheckEmail(ctx context.Context, req *gensts.CheckEmailReq) (resp *gensts.CheckEmailResp, err error) {
	resp = new(gensts.CheckEmailResp)
	code, err := s.Redis.GetCtx(ctx, fmt.Sprintf("%s:%s", consts.EmailCode, req.Email))
	if err != nil {
		return resp, err
	}

	if code != "" && code == req.Code {
		if err = s.Redis.SetexCtx(ctx, fmt.Sprintf("%s:%s", consts.PassCheckEmail, req.Email), "true", 300); err != nil {
			return resp, err
		}
		resp.Ok = true
	}
	return resp, nil
}

// 重置密码
func (s *AuthServiceImpl) SetPassword(ctx context.Context, req *gensts.SetPasswordReq) (resp *gensts.SetPasswordResp, err error) {
	resp = new(gensts.SetPasswordResp)
	var user *usermapper.User
	switch o := req.Key.(type) {
	case *gensts.SetPasswordReq_EmailOptions:
		value := ""
		if value, err = s.Redis.GetCtx(ctx, fmt.Sprintf("%s:%s", consts.PassCheckEmail, o.EmailOptions.Email)); err != nil {
			return resp, err
		}
		if value != "true" {
			return resp, consts.ErrNotPassEmailCheck
		}
		user, err = s.UserMongoMapper.FindOneByAuth(ctx, &usermapper.Auth{Type: int64(consts.EmailAuthType), AppId: o.EmailOptions.Email})
		if err != nil {
			return resp, err
		}
		_, err = s.UserMongoMapper.Update(ctx, &usermapper.User{ID: user.ID, PassWord: req.Password})
		if err != nil {
			return resp, err
		}

		if _, err = s.Redis.DelCtx(ctx, fmt.Sprintf("%s:%s", consts.PassCheckEmail, o.EmailOptions.Email)); err != nil {
			return resp, err
		}

	case *gensts.SetPasswordReq_UserIdOptions:
		user, err = s.UserMongoMapper.FindOne(ctx, o.UserIdOptions.UserId)
		if err != nil {
			return resp, err
		}
		if user.PassWord != o.UserIdOptions.Password {
			return resp, consts.ErrPasswordNotEqual
		}

		if _, err = s.UserMongoMapper.Update(ctx, &usermapper.User{ID: user.ID, PassWord: req.Password}); err != nil {
			return resp, err
		}
	}
	return resp, nil
}

// 发送邮件
func (s *AuthServiceImpl) SendEmail(ctx context.Context, req *gensts.SendEmailReq) (resp *gensts.SendEmailResp, err error) {
	resp = new(gensts.SendEmailResp)
	code, err := email.SendEmail(ctx, s.Config.EmailConf, req.Email, req.Subject)
	if err != nil {
		return resp, err
	}
	if err = s.Redis.SetexCtx(ctx, fmt.Sprintf("%s:%s", consts.EmailCode, req.Email), code, 300); err != nil {
		return resp, err
	}
	return resp, nil
}

// 注册
func (s *AuthServiceImpl) CreateAuth(ctx context.Context, req *gensts.CreateAuthReq) (resp *gensts.CreateAuthResp, err error) {
	resp = new(gensts.CreateAuthResp)
	auth := &usermapper.Auth{
		Type:       req.AuthType,
		AppId:      req.AppId,
		UnionId:    req.UnionId,
		PlatformId: req.PlatFormId,
	}
	_, err = s.UserMongoMapper.FindOneByAuth(ctx, auth)
	switch {
	case err == nil:
		return resp, consts.ErrHaveExist
	case errors.Is(err, consts.ErrNotFound):
		break
	default:
		return resp, err
	}

	password := req.Password
	if password == "" {
		password = consts.DefaultPassword
	}
	resp.UserId, err = s.UserMongoMapper.Insert(ctx, &usermapper.User{
		PassWord: password,
		Role:     req.Role,
		Auths:    []*usermapper.Auth{auth},
	})
	if err != nil {
		return resp, err
	}
	return resp, nil
}
