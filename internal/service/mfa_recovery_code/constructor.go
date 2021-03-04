package mfa_recovery_code

import (
	"github.com/nori-plugins/authentication/internal/config"
	"github.com/nori-plugins/authentication/internal/domain/helper/mfa_recovery_codes"
	"github.com/nori-plugins/authentication/internal/domain/repository"
	"github.com/nori-plugins/authentication/internal/domain/service"
)

type MfaRecoveryCodeService struct {
	mfaRecoveryCodeRepository repository.MfaRecoveryCodeRepository
	mfaRecoveryCodeHelper     mfa_recovery_codes.MfaRecoveryCodesHelper
	Config                    config.Config
}

type ServiceParams struct {
	MfaRecoveryCodeRepository repository.MfaRecoveryCodeRepository
	MfaRecoveryCodeHelper     mfa_recovery_codes.MfaRecoveryCodesHelper
	Config                    config.Config
}

func New(params ServiceParams) service.MfaRecoveryCodeService {
	return &MfaRecoveryCodeService{
		mfaRecoveryCodeRepository: params.MfaRecoveryCodeRepository,
		mfaRecoveryCodeHelper:     params.MfaRecoveryCodeHelper,
		Config:                    params.Config,
	}
}