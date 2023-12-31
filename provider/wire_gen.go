// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package provider

import (
	"github.com/CloudStriver/cloudmind-sts/biz/adaptor"
	"github.com/CloudStriver/cloudmind-sts/biz/application/service"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/config"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/mapper/user"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/stores/redis"
)

// Injectors from wire.go:

func NewStsServerImpl() (*adaptor.StsServerImpl, error) {
	configConfig, err := config.NewConfig()
	if err != nil {
		return nil, err
	}
	redisRedis := redis.NewRedis(configConfig)
	userMongoMapper := user.NewMongoMapper(configConfig)
	authServiceImpl := &service.AuthServiceImpl{
		Config:          configConfig,
		Redis:           redisRedis,
		UserMongoMapper: userMongoMapper,
	}
	stsServerImpl := &adaptor.StsServerImpl{
		Config:      configConfig,
		AuthService: authServiceImpl,
	}
	return stsServerImpl, nil
}
