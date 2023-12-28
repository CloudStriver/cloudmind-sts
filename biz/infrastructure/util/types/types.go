package types

import (
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/util/captcha/puzzle_captcha"
)

type CheckParams struct {
	Point *captcha.Point `json:"point"`
	Token string         `json:"token"`
}
type Token struct {
	AccessToken  string
	RefreshToken string
	ChatToken    string
}
type Data struct {
	Point *captcha.Point `json:"point"`
}
