package adaptor

import (
	"context"
	"github.com/CloudStriver/cloudmind-sts/biz/application/service"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/config"
	"github.com/CloudStriver/service-idl-gen-go/kitex_gen/cloudmind/sts"
)

type StsServerImpl struct {
	*config.Config
	AuthService service.AuthService
}

func (s *StsServerImpl) AppendAuth(ctx context.Context, req *sts.AppendAuthReq) (resp *sts.AppendAuthResp, err error) {
	return s.AuthService.AppendAuth(ctx, req)
}

func (s *StsServerImpl) Login(ctx context.Context, req *sts.LoginReq) (resp *sts.LoginResp, err error) {
	return s.AuthService.Login(ctx, req)
}

func (s *StsServerImpl) SetPassword(ctx context.Context, req *sts.SetPasswordReq) (res *sts.SetPasswordResp, err error) {
	return s.AuthService.SetPassword(ctx, req)
}

func (s *StsServerImpl) SendEmail(ctx context.Context, req *sts.SendEmailReq) (res *sts.SendEmailResp, err error) {
	return s.AuthService.SendEmail(ctx, req)
}

func (s *StsServerImpl) CheckEmail(ctx context.Context, req *sts.CheckEmailReq) (res *sts.CheckEmailResp, err error) {
	return s.AuthService.CheckEmail(ctx, req)
}

func (s *StsServerImpl) CreateAuth(ctx context.Context, req *sts.CreateAuthReq) (res *sts.CreateAuthResp, err error) {
	return s.AuthService.CreateAuth(ctx, req)
}
