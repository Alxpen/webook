package localsms

import (
	"context"
	"fmt"
	"strings"
)

// Service 是一个本地开发用的短信服务实现。
// 它不会真的发短信，而是把验证码打印到控制台，
// 方便本地调试时直接看到验证码。
type Service struct {
}

func NewService() *Service {
	return &Service{}
}

// Send 实现 sms.Service 接口
// tpl: 短信模板 ID
// args: 模板参数，验证码一般是 args[0]
// numbers: 手机号，支持一次发多个
func (s *Service) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	fmt.Printf("模拟发送短信, 模板ID: %s, 手机号: %s, 参数: %s\n",
		tpl, strings.Join(numbers, ","), strings.Join(args, ","))
	return nil
}
