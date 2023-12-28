package convertor

import (
	authmapper "github.com/CloudStriver/cloudmind-sts/biz/infrastructure/mapper/auth"
	gensts "github.com/CloudStriver/service-idl-gen-go/kitex_gen/cloudmind/sts"
)

func AuthToAuthMapper(u *gensts.Auth) *authmapper.Auth {
	return &authmapper.Auth{
		PassWord: u.Password,
		Role:     u.Role,
		Type:     u.Type,
		Key:      u.Key,
		UserId:   u.UserId,
	}
}
