package service

import (
	"context"
	"fmt"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/config"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/consts"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/convertor"
	usermapper "github.com/CloudStriver/cloudmind-sts/biz/infrastructure/mapper/user"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/util/email"
	"github.com/CloudStriver/go-pkg/utils/util/log"
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
	UserMongoMapper usermapper.UserMongoMapper
}

func (s *AuthServiceImpl) AppendAuth(ctx context.Context, req *gensts.AppendAuthReq) (resp *gensts.AppendAuthResp, err error) {
	resp = new(gensts.AppendAuthResp)
	if err = s.UserMongoMapper.AppendAuth(ctx, req.UserId, convertor.AuthToAuthMapper(req.AuthInfo)); err != nil {
		log.CtxError(ctx, "追加授权信息异常[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *AuthServiceImpl) Login(ctx context.Context, req *gensts.LoginReq) (resp *gensts.LoginResp, err error) {
	resp = new(gensts.LoginResp)
	user, err := s.UserMongoMapper.FindOneByAuth(ctx, convertor.AuthToAuthMapper(req.Auth))
	if errors.Is(err, consts.ErrNotFound) {
		return resp, nil
	}
	if err != nil {
		log.CtxError(ctx, "查询用户授权信息异常[%v]\n", err)
		return resp, err
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
		log.CtxError(ctx, "Redis获取缓存异常[%v]\n", err)
		return resp, err
	}

	if code != "" && code == req.Code {
		return resp, nil
	}

	return resp, consts.ErrCodeNotEqual
}

func (s *AuthServiceImpl) SetPassword(ctx context.Context, req *gensts.SetPasswordReq) (resp *gensts.SetPasswordResp, err error) {
	resp = new(gensts.SetPasswordResp)
	var user *usermapper.User
	switch o := req.Key.(type) {
	case *gensts.SetPasswordReq_EmailOptions:
		if _, err = s.CheckEmail(ctx, &gensts.CheckEmailReq{Email: o.EmailOptions.Email, Code: o.EmailOptions.Code}); err != nil {
			return resp, err
		}

		user, err = s.UserMongoMapper.FindOneByAuth(ctx, &usermapper.Auth{Type: int32(gensts.AuthType_email), AppId: o.EmailOptions.Email})
		if err != nil {
			log.CtxError(ctx, "查找授权信息异常[%v]\n", err)
			return resp, err
		}
		_, err = s.UserMongoMapper.Update(ctx, &usermapper.User{ID: user.ID, PassWord: req.Password})
		if err != nil {
			log.CtxError(ctx, "修改用户授权信息异常[%v]\n", err)
			return resp, err
		}
	case *gensts.SetPasswordReq_UserIdOptions:
		user, err = s.UserMongoMapper.FindOne(ctx, o.UserIdOptions.UserId)
		if err != nil {
			log.CtxError(ctx, "查找用户信息异常[%v]\n", err)
			return resp, err
		}
		if user.PassWord != o.UserIdOptions.Password {
			return resp, consts.ErrPasswordNotEqual
		}

		if _, err = s.UserMongoMapper.Update(ctx, &usermapper.User{ID: user.ID, PassWord: req.Password}); err != nil {
			log.CtxError(ctx, "修改用户授权信息异常[%v]\n", err)
			return resp, err
		}
	}
	return resp, nil
}

func (s *AuthServiceImpl) SendEmail(ctx context.Context, req *gensts.SendEmailReq) (resp *gensts.SendEmailResp, err error) {
	resp = new(gensts.SendEmailResp)
	code, err := email.SendEmail(ctx, s.Config.EmailConf, req.Email, req.Subject)
	if err != nil {
		log.CtxError(ctx, "发送邮件异常[%v]\n", err)
		return resp, err
	}
	if err = s.Redis.SetexCtx(ctx, fmt.Sprintf("%s:%s", consts.EmailCode, req.Email), code, 300); err != nil {
		return resp, err
	}
	return resp, nil
}

func (s *AuthServiceImpl) CreateAuth(ctx context.Context, req *gensts.CreateAuthReq) (resp *gensts.CreateAuthResp, err error) {
	resp = new(gensts.CreateAuthResp)
	if req.AuthInfo.AuthType == gensts.AuthType_email {
		_, err = s.CheckEmail(ctx, &gensts.CheckEmailReq{
			Email: req.AuthInfo.AppId,
			Code:  req.GetCode(),
		})
		if err != nil {
			return resp, err
		}
	}
	auth := convertor.AuthToAuthMapper(req.AuthInfo)
	_, err = s.UserMongoMapper.FindOneByAuth(ctx, auth)
	switch {
	case err == nil:
		return resp, consts.ErrHaveExist
	case errors.Is(err, consts.ErrNotFound):
		break
	default:
		return resp, err
	}
	resp.UserId, err = s.UserMongoMapper.Insert(ctx, &usermapper.User{
		PassWord: req.UserInfo.GetPassword(),
		Role:     int32(req.UserInfo.Role),
		Auths:    []*usermapper.Auth{auth},
	})
	if err != nil {
		log.CtxError(ctx, "插入用户授权信息异常[%v]\n", err)
		return resp, err
	}
	return resp, nil
}
