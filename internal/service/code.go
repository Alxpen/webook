package service

import (
	"context"
	"fmt"
	"math/rand"

	"webook/internal/repository"
	"webook/internal/service/sms"
)

const codeTplId = "1877556"

var (
	ErrCodeSendTooMany = repository.ErrCodeSendTooMany
	ErrCodeVerifyTooManyTimes = repository.ErrCodeVerifyTooManyTimes
)

type CodeService struct {
	repo   *repository.CodeRepository
	smsSvc sms.Service
}

func NewCodeService(repo *repository.CodeRepository, smsSvc sms.Service) *CodeService {
	return &CodeService{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

// Send 生成一个随机验证码，并发送
// biz 业务标识
func (svc *CodeService) Send(ctx context.Context, biz string, phone string) error {
	// 生成一个验证码
	code := svc.generateCode()

	// 塞进redis
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}

	// 发送出去
	err = svc.smsSvc.Send(ctx, codeTplId, []string{code}, phone)
	if err != nil {
		return err
	}
	return err
}

// Verify 验证验证码
func (svc *CodeService) Verify(ctx context.Context,
	biz string,
	phone string,
	inputCode string,
) (bool, error) {
	return svc.repo.Verify(ctx, biz, phone, inputCode)
}

func (svc *CodeService) generateCode() string {
	num := rand.Intn(1000000)
	return fmt.Sprintf("%06d", num)
}
