package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/config"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/consts"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/convertor"
	authmapper "github.com/CloudStriver/cloudmind-sts/biz/infrastructure/mapper/auth"
	captcha "github.com/CloudStriver/cloudmind-sts/biz/infrastructure/util/captcha/puzzle_captcha"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/util/types"
	"github.com/CloudStriver/go-pkg/utils/pconvertor"
	"github.com/CloudStriver/go-pkg/utils/uuid"
	gensts "github.com/CloudStriver/service-idl-gen-go/kitex_gen/cloudmind/sts"
	"github.com/bytedance/sonic"
	"github.com/google/wire"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/monc"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type AuthService interface {
	CreateCaptcha(ctx context.Context, _ *gensts.CreateCaptchaReq) (resp *gensts.CreateCaptchaResp, err error)
	CheckCaptcha(ctx context.Context, _ *gensts.CheckCaptchaReq) (resp *gensts.CheckCaptchaResp, err error)
	AddAuth(ctx context.Context, req *gensts.AddAuthReq) (resp *gensts.AddAuthResp, err error)
	CheckEmail(ctx context.Context, req *gensts.CheckEmailReq) (resp *gensts.CheckEmailResp, err error)
	SetPassword(ctx context.Context, req *gensts.SetPasswordReq) (resp *gensts.SetPasswordResp, err error)
	SendEmail(ctx context.Context, req *gensts.SendEmailReq) (resp *gensts.SendEmailResp, err error)
}

type AuthServiceImpl struct {
	Config         *config.Config
	Redis          *redis.Redis
	AuthMongMapper authmapper.AuthMongoMapper
}

func (s *AuthServiceImpl) CheckEmail(ctx context.Context, req *gensts.CheckEmailReq) (resp *gensts.CheckEmailResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (s *AuthServiceImpl) SetPassword(ctx context.Context, req *gensts.SetPasswordReq) (resp *gensts.SetPasswordResp, err error) {
	switch o := req.Key.(type) {
	case *gensts.SetPasswordReq_EmailOptions:
		checkResp, err := s.CheckEmail(ctx, &gensts.CheckEmailReq{Email: o.EmailOptions.Email})
		if err != nil {
			return resp, err
		}
		if checkResp.Error != "" {
			resp.Error = checkResp.Error
			return resp, nil
		}

		res, err := s.AuthMongMapper.UpdateByAuthKeyAndType(ctx, &authmapper.Auth{Key: o.EmailOptions.Email, Type: consts.Email, PassWord: req.Password})
		if err != nil {
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
			return resp, err
		}
		if auth.PassWord != o.UserIdOptions.Password {
			resp.Error = "密码错误"
			return resp, nil
		}

		if _, err = s.AuthMongMapper.Update(ctx, &authmapper.Auth{ID: auth.ID, PassWord: req.Password}); err != nil {
			return resp, err
		}
	}
	return resp, nil
}

func (s *AuthServiceImpl) SendEmail(ctx context.Context, req *gensts.SendEmailReq) (resp *gensts.SendEmailResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (s *AuthServiceImpl) AddAuth(ctx context.Context, req *gensts.AddAuthReq) (resp *gensts.AddAuthResp, err error) {
	resp = new(gensts.AddAuthResp)
	if _, err = s.AuthMongMapper.Insert(ctx, convertor.AuthToAuthMapper(req.Auth)); err != nil {
		return nil, err
	}
	return resp, nil
}

var AuthSet = wire.NewSet(
	wire.Struct(new(AuthServiceImpl), "*"),
	wire.Bind(new(AuthService), new(*AuthServiceImpl)),
)

func (s *AuthServiceImpl) CreateCaptcha(ctx context.Context, _ *gensts.CreateCaptchaReq) (resp *gensts.CreateCaptchaResp, err error) {
	resp = new(gensts.CreateCaptchaResp)
	if err = captcha.LoadBackgroudImages("./biz/infrastructure/util/captcha/images/puzzle_captcha/backgroud"); err != nil {
		logx.Errorf("验证码背景加载异常[%v]\n", err)
		return resp, err
	}
	if err = captcha.LoadBlockImages("./biz/infrastructure/util/captcha/images/puzzle_captcha/block"); err != nil {
		logx.Errorf("验证码图片加载异常[%v]\n", err)
		return resp, err
	}

	ret, err := captcha.Run()
	if err != nil {
		logx.Errorf("验证码生成异常[%v]\n", err)
		return resp, err
	}

	resp.Key = uuid.Newuuid()
	resp.OriginalImageBase64 = ret.BackgroudImg
	resp.JigsawImageBase64 = ret.BlockImg
	if err = s.Redis.SetexCtx(ctx, fmt.Sprintf("%s:%s", consts.CaptchaKey, resp.Key), pconvertor.StructToJsonString(&types.CheckParams{Point: ret.Point}), 60); err != nil {
		logx.Errorf("Redis设置缓存异常[%v]\n", err)
	}

	return resp, nil
}

func (s *AuthServiceImpl) CheckCaptcha(ctx context.Context, req *gensts.CheckCaptchaReq) (resp *gensts.CheckCaptchaResp, err error) {
	resp = new(gensts.CheckCaptchaResp)
	value, err := s.Redis.GetCtx(ctx, fmt.Sprintf("%s:%s", consts.CaptchaKey, req.Key))
	if err != nil {
		logx.Errorf("Redis获取缓存异常[%v]\n", err)
	}
	if value == "" {
		resp.Error = "验证码过期"
		return resp, nil
	}
	var c *captcha.Point
	if err = sonic.Unmarshal([]byte(value), c); err != nil {
		return resp, err
	}
	if err = captcha.Check(&captcha.Point{X: int(req.Point.X), Y: int(req.Point.Y)}, c); err != nil {
		resp.Error = "验证码错误"
		return resp, nil
	}
	return resp, nil
}
