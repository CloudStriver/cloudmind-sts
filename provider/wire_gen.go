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
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/util/filter"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/util/sdk/cos"
)

// Injectors from wire.go:

func NewStsServerImpl() (*adaptor.StsServerImpl, error) {
	configConfig, err := config.NewConfig()
	if err != nil {
		return nil, err
	}
	redisRedis := redis.NewRedis(configConfig)
	iUserMongoMapper := user.NewMongoMapper(configConfig)
	authServiceImpl := &service.AuthServiceImpl{
		Config:          configConfig,
		Redis:           redisRedis,
		UserMongoMapper: iUserMongoMapper,
	}
	cosSDK, err := cos.NewCosSDK(configConfig)
	if err != nil {
		return nil, err
	}
	cosService := service.CosService{
		Config: configConfig,
		CosSDK: cosSDK,
	}
	illegalWordsSearch := filter.NewFilter(configConfig)
	filterService := service.FilterService{
		Config: configConfig,
		Filter: illegalWordsSearch,
	}
	stsServerImpl := &adaptor.StsServerImpl{
		Config:        configConfig,
		AuthService:   authServiceImpl,
		CosService:    cosService,
		FilterService: filterService,
	}
	return stsServerImpl, nil
}
