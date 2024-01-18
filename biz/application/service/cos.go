package service

import (
	"context"
	"fmt"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/config"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/util/sdk/cos"
	gensts "github.com/CloudStriver/service-idl-gen-go/kitex_gen/cloudmind/sts"
	"github.com/google/wire"
	cossts "github.com/tencentyun/qcloud-cos-sts-sdk/go"
	"github.com/xh-polaris/platform-sts/biz/infrastructure/consts"
	"time"
)

type ICosService interface {
	GenCosSts(ctx context.Context, req *gensts.GenCosStsReq) (*gensts.GenCosStsResp, error)
	GenSignedUrl(ctx context.Context, req *gensts.GenSignedUrlReq) (*gensts.GenSignedUrlResp, error)
	DeleteObject(ctx context.Context, req *gensts.DeleteObjectReq) (*gensts.DeleteObjectResp, error)
}

type CosService struct {
	Config *config.Config
	CosSDK *cos.CosSDK
}

var CosSet = wire.NewSet(
	wire.Struct(new(CosService), "*"),
	wire.Bind(new(ICosService), new(*CosService)),
)

func (s *CosService) GenCosSts(ctx context.Context, req *gensts.GenCosStsReq) (*gensts.GenCosStsResp, error) {
	cosConfig := s.Config.CosConfig
	stsOption := &cossts.CredentialOptions{
		// 临时密钥有效时长，单位是秒
		DurationSeconds: int64(10 * time.Minute.Seconds()),
		Region:          cosConfig.Region,
		Policy: &cossts.CredentialPolicy{
			Statement: []cossts.CredentialPolicyStatement{
				{
					// 密钥的权限列表。简单上传和分片需要以下的权限，其他权限列表请看 https://cloud.tencent.com/document/product/436/31923
					Action: []string{
						// 简单上传
						"name/cos:PostObject",
						"name/cos:PutObject",
						// 分片上传
						"name/cos:InitiateMultipartUpload",
						"name/cos:ListMultipartUploads",
						"name/cos:ListParts",
						"name/cos:UploadPart",
						"name/cos:CompleteMultipartUpload",
					},
					Effect: "allow",
					// 密钥可控制的资源列表。此处开放名字为用户ID的文件夹及其子文件夹
					Resource: []string{
						fmt.Sprintf("qcs::cos:%s:uid/%s:%s/%s",
							cosConfig.Region, cosConfig.AppId, cosConfig.BucketName, req.Path),
					},
				},
			},
		},
	}

	res, err := s.CosSDK.GetCredential(ctx, stsOption)
	if err != nil {
		return nil, err
	}

	return &gensts.GenCosStsResp{
		SecretId:     res.Credentials.TmpSecretID,
		SecretKey:    res.Credentials.TmpSecretKey,
		SessionToken: res.Credentials.SessionToken,
		ExpiredTime:  int64(res.ExpiredTime),
		StartTime:    int64(res.StartTime),
	}, nil
}

func (s *CosService) GenSignedUrl(ctx context.Context, req *gensts.GenSignedUrlReq) (*gensts.GenSignedUrlResp, error) {
	signedUrl, err := s.CosSDK.GetPresignedURL(ctx, req.Method, req.Path, req.SecretId, req.SecretKey, time.Minute, nil)
	if err != nil {
		return nil, err
	}
	return &gensts.GenSignedUrlResp{SignedUrl: signedUrl.String()}, nil
}

func (s *CosService) DeleteObject(ctx context.Context, req *gensts.DeleteObjectReq) (*gensts.DeleteObjectResp, error) {
	res, err := s.CosSDK.Delete(ctx, req.Path)
	if err != nil || res.StatusCode != 200 {
		return nil, consts.ErrCannotDeleteObject
	}
	return &gensts.DeleteObjectResp{}, nil
}
