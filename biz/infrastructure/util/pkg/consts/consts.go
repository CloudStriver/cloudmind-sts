package consts

const (
	Score             = "_score"
	CreateAt          = "createAt"
	ConfirmEmailCode  = "ConfirmEmailCode"
	DefaultAvatarUrl  = "d2042520dce2223751906a11e547d43e.png"
	UserAuthTypeEmail = 0
	CaptchaKey        = "CaptchaKey"
	Name              = "name"
	UpdateAt          = "updateAt"
	UserId            = "_id"
)

const EmailTemplate = `
	<!DOCTYPE html>
	<html lang="en">
	<head>
	  <meta charset="UTF-8">
	  <meta name="viewport" content="width=device-width, initial-scale=1.0">
	  <title>重置密码验证码</title>
	  <style>
	      body {
	          font-family: Arial, sans-serif;
	          background-color: #f0f0f0;
	          margin: 0;
	          padding: 0;
	      }
	      .container {
	          max-width: 600px;
	          margin: 0 auto;
	          padding: 20px;
	          background-color: #ffffff;
	          border-radius: 5px;
	          box-shadow: 0px 0px 10px rgba(0, 0, 0, 0.1);
	      }
	      p {
	          font-size: 16px;
	          line-height: 1.6;
	          color: #333333;
	      }
	      strong {
	          font-weight: bold;
	      }
	  </style>
	</head>
	<body>
	<div class="container">
	  <p>你好，</p>
	  <p>你此次重置密码的验证码如下，请在 5 分钟内输入验证码进行下一步操作。如非你本人操作，请忽略此邮件。</p>
	  <p><strong>验证码：</strong>{{.code}}</p>
	</div>
	</body>
	</html>
	`
