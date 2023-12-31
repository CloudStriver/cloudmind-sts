package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/config"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/consts"
	authmapper "github.com/CloudStriver/cloudmind-sts/biz/infrastructure/mapper/auth"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/util/captcha/puzzle_captcha"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/util/email"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/util/types"
	"github.com/CloudStriver/go-pkg/utils/pconvertor"
	"github.com/CloudStriver/go-pkg/utils/util/log"
	"github.com/CloudStriver/go-pkg/utils/uuid"
	gensts "github.com/CloudStriver/service-idl-gen-go/kitex_gen/cloudmind/sts"
	"github.com/bytedance/sonic"
	"github.com/google/wire"
	"github.com/zeromicro/go-zero/core/stores/monc"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthService interface {
	CreateCaptcha(ctx context.Context, _ *gensts.CreateCaptchaReq) (resp *gensts.CreateCaptchaResp, err error)
	CheckCaptcha(ctx context.Context, _ *gensts.CheckCaptchaReq) (resp *gensts.CheckCaptchaResp, err error)
	CreateAuth(ctx context.Context, req *gensts.CreateAuthReq) (resp *gensts.CreateAuthResp, err error)
	CheckEmail(ctx context.Context, req *gensts.CheckEmailReq) (resp *gensts.CheckEmailResp, err error)
	SetPassword(ctx context.Context, req *gensts.SetPasswordReq) (resp *gensts.SetPasswordResp, err error)
	SendEmail(ctx context.Context, req *gensts.SendEmailReq) (resp *gensts.SendEmailResp, err error)
	Login(ctx context.Context, req *gensts.LoginReq) (resp *gensts.LoginResp, err error)
}

type AuthServiceImpl struct {
	Config         *config.Config
	Redis          *redis.Redis
	AuthMongMapper authmapper.AuthMongoMapper
}

func (s *AuthServiceImpl) Login(ctx context.Context, req *gensts.LoginReq) (resp *gensts.LoginResp, err error) {
	resp = new(gensts.LoginResp)
	auth, err := s.AuthMongMapper.FindOneByAuthKeyAndType(ctx, req.Email, consts.Email)
	switch {
	case errors.Is(err, monc.ErrNotFound):
		resp.Error = "邮箱不存在"
		return resp, nil
	case err == nil:
		if auth.PassWord != req.Password {
			resp.Error = "密码错误"
		} else {
			resp.UserId = auth.UserId
		}
		return resp, nil
	default:
		log.CtxError(ctx, "查询用户密码异常[%v]\n", err)
		return resp, err
	}
}

var AuthSet = wire.NewSet(
	wire.Struct(new(AuthServiceImpl), "*"),
	wire.Bind(new(AuthService), new(*AuthServiceImpl)),
)

func (s *AuthServiceImpl) CheckEmail(ctx context.Context, req *gensts.CheckEmailReq) (resp *gensts.CheckEmailResp, err error) {
	resp = new(gensts.CheckEmailResp)
	code, err := s.Redis.GetCtx(ctx, fmt.Sprintf("%s:%s", consts.EmailCode, req.Email))
	if err != nil {
		log.CtxError(ctx, "Redis获取缓存异常[%v]\n", err)
		return resp, err
	}

	if code == "" {
		resp.Error = "验证码已过期"
	} else if code != req.Code {
		resp.Error = "验证码错误"
	}

	return resp, nil
}

func (s *AuthServiceImpl) SetPassword(ctx context.Context, req *gensts.SetPasswordReq) (resp *gensts.SetPasswordResp, err error) {
	switch o := req.Key.(type) {
	case *gensts.SetPasswordReq_EmailOptions:
		checkResp, err := s.CheckEmail(ctx, &gensts.CheckEmailReq{Email: o.EmailOptions.Email})
		if err != nil {
			log.CtxError(ctx, "调用CheckEmail异常[%v]\n", err)
			return resp, err
		}
		if checkResp.Error != "" {
			resp.Error = checkResp.Error
			return resp, nil
		}

		res, err := s.AuthMongMapper.UpdateByAuthKeyAndType(ctx, &authmapper.Auth{Key: o.EmailOptions.Email, Type: consts.Email, PassWord: req.Password})
		if err != nil {
			log.CtxError(ctx, "修改用户密码异常[%v]\n", err)
			return resp, err
		}

		if res.ModifiedCount == 0 {
			resp.Error = "邮箱不存在"
			return resp, nil
		}
	case *gensts.SetPasswordReq_UserIdOptions:
		auth, err := s.AuthMongMapper.FindOneByUserId(ctx, o.UserIdOptions.UserId)
		switch {
		case errors.Is(err, monc.ErrNotFound):
			resp.Error = "用户不存在"
			return resp, nil
		case err == nil:
		default:
			log.CtxError(ctx, "查找授权信息异常[%v]\n", err)
			return resp, err
		}
		if auth.PassWord != o.UserIdOptions.Password {
			resp.Error = "密码错误"
			return resp, nil
		}

		if _, err = s.AuthMongMapper.Update(ctx, &authmapper.Auth{ID: auth.ID, PassWord: req.Password}); err != nil {
			log.CtxError(ctx, "修改用户密码异常[%v]\n", err)
			return resp, err
		}
	}
	return resp, nil
}

func (s *AuthServiceImpl) SendEmail(ctx context.Context, req *gensts.SendEmailReq) (resp *gensts.SendEmailResp, err error) {
	resp = new(gensts.SendEmailResp)
	code, err := email.SendEmail(s.Config.EmailConf, req.Email, req.Subject)
	if err != nil {
		log.CtxError(ctx, "发送邮件异常[%v]\n", err)
		return resp, err
	}

	if err = s.Redis.SetexCtx(ctx, fmt.Sprintf("%s:%s", consts.EmailCode, req.Email), code, 60); err != nil {
		log.CtxError(ctx, "Redis设置缓存异常[%v]\n", err)
	}

	return resp, nil
}

func (s *AuthServiceImpl) CreateAuth(ctx context.Context, req *gensts.CreateAuthReq) (resp *gensts.CreateAuthResp, err error) {
	resp = new(gensts.CreateAuthResp)
	auth := &authmapper.Auth{
		PassWord: req.Password,
		Role:     int32(req.Role),
		Type:     int32(req.Type),
		Key:      req.Key,
		UserId:   req.UserId,
	}
	_, err = s.AuthMongMapper.Insert(ctx, auth)
	switch {
	case mongo.IsDuplicateKeyError(err):
		resp.Error = "已经注册过"
		return resp, nil
	case err != nil:
		log.CtxError(ctx, "新增授权信息异常[%v]\n", err)
		return resp, err
	}
	return resp, nil
}

func (s *AuthServiceImpl) CreateCaptcha(ctx context.Context, _ *gensts.CreateCaptchaReq) (resp *gensts.CreateCaptchaResp, err error) {
	resp = new(gensts.CreateCaptchaResp)
	if err = captcha.LoadBackgroudImages("./biz/infrastructure/util/captcha/images/puzzle_captcha/backgroud"); err != nil {
		log.CtxError(ctx, "验证码背景加载异常[%v]\n", err)
		return resp, err
	}
	if err = captcha.LoadBlockImages("./biz/infrastructure/util/captcha/images/puzzle_captcha/block"); err != nil {
		log.CtxError(ctx, "验证码图片加载异常[%v]\n", err)
		return resp, err
	}

	ret, err := captcha.Run()
	if err != nil {
		log.CtxError(ctx, "验证码生成异常[%v]\n", err)
		return resp, err
	}

	resp.Key = uuid.Newuuid()
	resp.OriginalImageBase64 = ret.BackgroudImg
	resp.JigsawImageBase64 = ret.BlockImg
	if err = s.Redis.SetexCtx(ctx, fmt.Sprintf("%s:%s", consts.CaptchaKey, resp.Key), pconvertor.StructToJsonString(&types.CheckParams{Point: ret.Point}), 60); err != nil {
		log.CtxError(ctx, "Redis设置缓存异常[%v]\n", err)
	}

	return resp, nil
}

func (s *AuthServiceImpl) CheckCaptcha(ctx context.Context, req *gensts.CheckCaptchaReq) (resp *gensts.CheckCaptchaResp, err error) {
	resp = new(gensts.CheckCaptchaResp)
	value, err := s.Redis.GetCtx(ctx, fmt.Sprintf("%s:%s", consts.CaptchaKey, req.Key))
	if err != nil {
		log.CtxError(ctx, "Redis获取缓存异常[%v]\n", err)
	}
	if value == "" {
		resp.Error = "验证码过期"
		return resp, nil
	}
	var c *captcha.Point
	if err = sonic.Unmarshal([]byte(value), c); err != nil {
		log.CtxError(ctx, "sonic.Unmarshal异常[%v]\n", err)
		return resp, err
	}
	if err = captcha.Check(&captcha.Point{X: int(req.Point.X), Y: int(req.Point.Y)}, c); err != nil {
		resp.Error = "验证码错误"
		return resp, nil
	}
	return resp, nil
}
