package config

import "github.com/nori-io/common/v5/pkg/domain/config"

type Config struct {
	CookiesPath                   config.String
	CookiesDomain                 config.String
	CookiesExpires                config.Int64
	CookiesMaxAge                 config.Int
	CookiesSecure                 config.Bool
	CookiesHttpOnly               config.Bool
	CookiesSameSite               config.Int
	CookiesName                   config.String
	EmailVerification             config.Bool
	EmailActivationCodeTTLSeconds config.Int64
	MfaRecoveryCodePattern        config.String
	MfaRecoveryCodeSymbols        config.String
	MfaRecoveryCodeLength         config.Int
	MfaRecoveryCodeCount          config.Int
	Issuer                        config.String
	PasswordBcryptCost            config.Int
	UrlPrefix                     config.String
	UrlLogoutRedirect             config.String
}
