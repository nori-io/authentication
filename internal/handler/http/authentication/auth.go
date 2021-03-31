package authentication

import (
	"net/http"
	"time"

	"github.com/nori-plugins/authentication/internal/domain/helper/cookie"
	"github.com/nori-plugins/authentication/pkg/enum/session_status"

	s "github.com/nori-io/interfaces/nori/session"

	"github.com/nori-plugins/authentication/internal/config"

	"github.com/nori-plugins/authentication/internal/handler/http/response"

	"github.com/nori-io/common/v4/pkg/domain/logger"

	"github.com/nori-plugins/authentication/internal/domain/entity"

	"github.com/nori-plugins/authentication/internal/domain/service"
)

type AuthenticationHandler struct {
	authenticationService service.AuthenticationService
	logger                logger.FieldLogger
	config                config.Config
	session               s.Session
	cookieHelper          cookie.CookieHelper
}

type Params struct {
	AuthenticationService service.AuthenticationService
	Logger                logger.FieldLogger
	Config                config.Config
	Session               s.Session
	CookieHelper          cookie.CookieHelper
}

func New(params Params) *AuthenticationHandler {
	return &AuthenticationHandler{
		authenticationService: params.AuthenticationService,
		logger:                params.Logger,
		config:                params.Config,
		session:               params.Session,
		cookieHelper:          params.CookieHelper,
	}
}

func (h *AuthenticationHandler) Session(w http.ResponseWriter, r *http.Request) {
	sessionId, err := h.cookieHelper.GetSessionID(r)
	if err != nil {
		h.logger.Error("%s", err)
		http.Error(w, http.ErrNoCookie.Error(), http.StatusUnauthorized)
	}

	err = h.session.Get([]byte(sessionId), session_status.Active)
	if err != nil {
		h.logger.Error("%s", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}

	sess, user, err := h.authenticationService.GetSessionInfo(r.Context(), sessionId)
	if err != nil {
		h.logger.Error("%s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)

	response.JSON(w, r, SessionResponse{
		Success:  true,
		Message:  "session exists",
		Email:    user.Email,
		Phone:    user.PhoneCountryCode + user.PhoneNumber,
		OpenedAt: sess.OpenedAt,
	})
}

func (h *AuthenticationHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	data, err := newSignUpData(r)
	if err != nil {
		h.logger.Error("%s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	//@todo проверить, нет ли у пользователя сессии, если есть, то выдать ошибку 403
	//@todo если пользователь уже зарегистрирован, то выдать ошибку 409
	_, err = h.authenticationService.SignUp(r.Context(), data)
	if err != nil {
		h.logger.Error("%s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *AuthenticationHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	data, err := newSignInData(r)
	if err != nil {
		h.logger.Error("%s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	sess, mfaType, err := h.authenticationService.SignIn(r.Context(), data)
	if err != nil {
		h.logger.Error("%s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	h.cookieHelper.SetSession(w, sess)
	//@todo логировать положительные действия?
	w.WriteHeader(http.StatusOK)

	response.JSON(w, r, SignInResponse{
		Success: true,
		Message: "User sign in",
		MfaType: *mfaType,
	})
}

func (h *AuthenticationHandler) SignInMfa(w http.ResponseWriter, r *http.Request) {
	sessionId, err := r.Cookie("ssid")
	if err != nil {
		h.logger.Error("%s", err)
		http.Error(w, http.ErrNoCookie.Error(), http.StatusUnauthorized)
	}

	h.session.Get([]byte(sessionId.Value), session_status.Active)
	if err != nil {
		h.logger.Error("%s", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}

	data, err := newSignInMfaData(r)
	if err != nil {
		h.logger.Error("%s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	sess, err := h.authenticationService.SignInMfa(r.Context(), data)
	if err != nil {
		h.logger.Error("%s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	h.cookieHelper.SetSession(w, sess)

	response.JSON(w, r, SignInMfaResponse{
		Success: true,
		Message: "User sign in by mfa",
	})
}

func (h *AuthenticationHandler) SignOut(w http.ResponseWriter, r *http.Request) {
	sessionId, err := r.Cookie("ssid")
	if err != nil {
		h.logger.Error("%s", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}

	data := &entity.Session{
		ID:         0,
		UserID:     0,
		SessionKey: nil,
		Status:     0,
		OpenedAt:   time.Time{},
		ClosedAt:   time.Time{},
		UpdatedAt:  time.Time{},
	}

	err = h.session.Get([]byte(sessionId.Value), data)
	if err != nil {
		h.logger.Error("%s", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}

	if data.Status != session_status.Active {

	}

	if err := h.authenticationService.SignOut(r.Context(), &entity.Session{SessionKey: []byte(sessionId.Value)}); err != nil {
		h.logger.Error("%s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// todo: redirect

	http.Redirect(w, r, "/", 0)
}
