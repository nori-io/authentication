package authentication

import (
	"context"
	"time"
)

type Social interface {
	GetAccounts(ctx context.Context, userID uint64) ([]SocialAccount, error)
	GetAccountsByFilter(ctx context.Context) ([]SocialAccount, error)
	GetAccessToken(ctx context.Context, refreshToken string, serviceProviderID uint64) (accessToken string, err error)
	GetServiceProviders(ctx context.Context) ([]ServiceProvider, error)
}

type SocialFilter struct {
	ID                  uint64
	ExternalID          string
	FirstName           string
	LastName            string
	FullName            string
	Email               string
	ServiceProviderName string
	Offset              int
	Limit               int
}

type SocialAccount struct {
	ID              uint64
	User_ID         uint64
	ExternalID      string
	FirstName       string
	LastName        string
	FullName        string
	Email           string
	AvatarURL       string
	ServiceProvider ServiceProvider
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ServiceProvider struct {
	ID        uint64
	Name      string
	Logo      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
