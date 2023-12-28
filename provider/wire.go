//go:build wireinject
// +build wireinject

package provider

import (
	"github.com/CloudStriver/cloudmind-sts/biz/adaptor"
	"github.com/google/wire"
)

func NewStsServerImpl() (*adaptor.StsServerImpl, error) {
	wire.Build(
		wire.Struct(new(adaptor.StsServerImpl), "*"),
		AllProvider,
	)
	return nil, nil
}
