package convertor

import (
	authmapper "github.com/CloudStriver/cloudmind-sts/biz/infrastructure/mapper/user"
	gensts "github.com/CloudStriver/service-idl-gen-go/kitex_gen/cloudmind/sts"
)

func AuthToAuthMapper(u *gensts.AuthInfo) *authmapper.Auth {
	return &authmapper.Auth{
		Type:    int32(u.AuthType),
		AppId:   u.AppId,
		UnionId: u.UnionId,
	}
}
