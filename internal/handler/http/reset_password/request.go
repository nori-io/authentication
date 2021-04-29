package reset_password

import (
	"encoding/json"
	"net/http"

	"github.com/nori-plugins/authentication/internal/domain/service"
)

type RestorePasswordRequest struct {
	Email string `json:"email"`
}

func newRequestResetPasswordEmailData(r *http.Request) (service.RequestResetPasswordEmailData, error) {
	var body RestorePasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return service.RequestResetPasswordEmailData{}, err
	}
	return service.RequestResetPasswordEmailData{
		Email: body.Email,
	}, nil
}

type RestorePasswordSetRequest struct {
	token    string `json:"token"`
	password string `json:"password"`
}

func NewSetNewPasswordByRestorePasswordEmailToken(r *http.Request) (service.SetNewPasswordByRestorePasswordEmailTokenData, error) {
	var body RestorePasswordSetRequest

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return service.SetNewPasswordByRestorePasswordEmailTokenData{}, err
	}
	return service.SetNewPasswordByRestorePasswordEmailTokenData{}, nil
}