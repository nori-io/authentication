package service

import (
	"context"
	"time"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/nori-plugins/authentication/pkg/enum/users_action"
)

type UserLogService interface {
	Create(ctx context.Context, data UserLogCreateData) error
}

type UserLogCreateData struct {
	UserID    uint64
	Action    users_action.Action
	SessionID uint64
	Meta      string
	CreatedAt time.Time
}

func (d UserLogCreateData) Validate() error {
	return v.Errors{
		"user_ID":    v.Validate(d.UserID, v.Required),
		"action":     v.Validate(d.Action, v.Required),
		"session_ID": v.Validate(d.UserID, v.Required),
		"meta":       v.Validate(d.Action, v.Required),
		"created_at": v.Validate(d.CreatedAt, v.Required),
	}.Filter()
}
