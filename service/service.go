package service

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"reflect"
	"time"

	rest "github.com/cheebo/gorest"
	"github.com/cheebo/rand"
	"github.com/dgrijalva/jwt-go"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/nori-io/nori-common/logger"
	"github.com/nori-io/nori-interfaces/interfaces"

	"github.com/nori-io/authentication/service/database"
)

type Service interface {
	SignUp(ctx context.Context, req SignUpRequest, parameters PluginParameters) (resp *SignUpResponse)
	SignIn(ctx context.Context, req SignInRequest, parameters PluginParameters) (resp *SignInResponse)
	SignOut(ctx context.Context, req SignOutRequest) (resp *SignOutResponse)
	RecoveryCodes(ctx context.Context, req RecoveryCodesRequest) (resp *RecoveryCodesResponse)
	/*SignInSocial(ctx context.Context, req http.Request, parameters PluginParameters) (resp *SignInSocialResponse)
	SignOutSocial(res http.ResponseWriter, req *http.Request)*/
}

type Config struct {
	Sub                                func() string
	Iss                                func() string
	UserType                           func() []interface{}
	UserTypeDefault                    func() string
	UserRegistrationByPhoneNumber      func() bool
	UserRegistrationByEmailAddress     func() bool
	UserMfaType                        func() string
	ActivationTimeForActivationMinutes func() uint
	ActivationCode                     func() bool
	Oath2ProvidersVKClientKey          func() string
	Oath2ProvidersVKClientSecret       func() string
	Oath2ProvidersVKRedirectUrl       func() string
}

type service struct {
	auth    interfaces.Auth
	cache   interfaces.Cache
	cfg     *Config
	db      database.Database
	log     logger.Writer
	mail    interfaces.Mail
	session interfaces.Session
}

type sessionData struct {
	name string
}

func NewService(
	auth interfaces.Auth,
	cache interfaces.Cache,
	cfg *Config,
	db database.Database,
	log logger.Writer,
	mail interfaces.Mail,
	session interfaces.Session,
) Service {
	return &service{
		auth:    auth,
		cache:   cache,
		cfg:     cfg,
		db:      db,
		log:     log,
		mail:    mail,
		session: session,
	}
}

func (s *service) SignUp(ctx context.Context, req SignUpRequest, parameters PluginParameters) (resp *SignUpResponse) {

	var err error
	var modelAuth *database.AuthModel
	var modelUsers *database.UsersModel
	resp = &SignUpResponse{}

	errField := rest.ErrFieldResp{
		Meta: rest.ErrMeta{
			ErrCode: 0,
		},
	}
	if len(req.Email) != 0 {
		modelAuth, err = s.db.Auth().FindByEmail(req.Email)
	} else if len(req.PhoneCountryCode+req.PhoneCountryCode) != 0 {
		modelAuth, err = s.db.Auth().FindByPhone(req.PhoneCountryCode, req.PhoneNumber)
	}

	if modelAuth != nil && modelAuth.Id != 0 {
		resp.Email = req.Email
		resp.PhoneCountryCode = req.PhoneCountryCode
		resp.PhoneNumber = req.PhoneNumber
		errField.AddError("phone, email", 400, "User already exists.")
	}

	if err != nil {
		resp.Err = err
		resp.Email = req.Email
		resp.PhoneCountryCode = req.PhoneCountryCode
		resp.PhoneNumber = req.PhoneNumber

		return resp
	}

	if errField.HasErrors() {

		resp.Err = errField
		return resp
	}

	modelAuth = &database.AuthModel{
		Email:            req.Email,
		Password:         []byte(req.Password),
		PhoneCountryCode: req.PhoneCountryCode,
		PhoneNumber:      req.PhoneNumber,
	}

	modelUsers = &database.UsersModel{
		Type:     req.Type,
		Mfa_type: req.MfaType,
	}

	if parameters.ActivationCode {
		modelUsers.Status_account = "locked"
	} else {
		modelUsers.Status_account = "active"
	}
	err = s.db.Users().Create(modelAuth, modelUsers)
	if err != nil {
		s.log.Error(err)
		resp.Err = rest.ErrFieldResp{
			Meta: rest.ErrMeta{
				ErrCode:    500,
				ErrMessage: err.Error(),
			},
		}

		return resp
	}

	resp.Email = req.Email
	resp.PhoneCountryCode = req.PhoneCountryCode
	resp.PhoneNumber = req.PhoneNumber

	return resp
}

func (s *service) SignIn(ctx context.Context, req SignInRequest, parameters PluginParameters) (resp *SignInResponse) {
	resp = &SignInResponse{}
	var model, modelFindByEmail, modelFindByPhone *database.AuthModel
	var errFindByEmail, errFindPhone error

	if parameters.UserRegistrationByEmailAddress {

		modelFindByEmail, errFindByEmail = s.db.Auth().FindByEmail(req.Name)
		if errFindByEmail != nil {
			resp.User.UserName = req.Name
			resp.Err = errFindByEmail
			return resp
		}
		if modelFindByEmail.Id != 0 {
			model = modelFindByEmail
		}
	}

	if parameters.UserRegistrationByPhoneNumber {
		modelFindByPhone, errFindPhone = s.db.Auth().FindByPhone(req.Name, "")
		if errFindPhone != nil {
			resp.User.UserName = req.Name
			resp.Err = errFindPhone
			return resp
		}
		if modelFindByPhone.Id != 0 {
			model = modelFindByPhone
		}
	}

	if model == nil {
		resp.User.UserName = req.Name
		resp.Err = rest.ErrResp{Meta: rest.ErrMeta{ErrMessage: "User not found", ErrCode: 0}}
		return resp
	}
	var userId uint64
	userId = model.Id
	result, err := database.VerifyPassword([]byte(req.Password), model.Salt, model.Password)

	if (!result) || (err != nil) {
		resp.Id = userId
		resp.User.UserName = req.Name
		resp.Err = rest.ErrResp{Meta: rest.ErrMeta{ErrMessage: "Incorrect Password", ErrCode: 0}}

		return resp
	}

	modelAuthenticationHistory := &database.AuthenticationHistoryModel{
		UserId: userId,
	}

	err = s.db.AuthenticationHistory().Create(modelAuthenticationHistory)
	if err != nil {
		s.log.Error(err)
		resp.User.UserName = req.Name
		resp.Err = rest.ErrFieldResp{
			Meta: rest.ErrMeta{
				ErrCode:    500,
				ErrMessage: err.Error(),
			},
		}
		return resp
	}

	sid := rand.RandomAlphaNum(32)

	token, err := s.auth.AccessToken(func(op interface{}) interface{} {
		key, ok := op.(string)
		if !ok || key == "" {
			return ""
		}
		switch key {
		case "raw":
			return map[string]string{
				"id":   string(userId),
				"name": req.Name,
			}
		case "jti":
			return sid
		case "sub":
			return s.cfg.Sub()
		case "iss":
			return s.cfg.Iss()
		default:
			return ""
		}
	})

	if err != nil {
		resp.Err = rest.ErrorInternal(err.Error())
		return resp
	}
	s.session.Save([]byte(sid), sessionData{name: req.Name}, 0)

	resp.Id = uint64(userId)
	resp.Token = token

	if model.Id != 0 {
		resp.User = UserResponse{UserName: req.Name}
	}

	return resp
}

func (s *service) SignOut(ctx context.Context, req SignOutRequest) (resp *SignOutResponse) {

	resp = &SignOutResponse{}

	value := ctx.Value("nori.auth.data")

	var name string

	if val, ok := value.(jwt.MapClaims)["raw"]; ok {
		reflect.TypeOf(val)
		if val2, ok2 := val.(map[string]interface{})["name"]; ok2 {
			name = fmt.Sprint(val2)
		}

	}

	req = SignOutRequest{}
	modelFindEmail, errFindEmail := s.db.Auth().FindByEmail(name)
	modelFindPhone, errFindPhone := s.db.Auth().FindByPhone(name, "")
	if (errFindEmail != nil) && (errFindPhone != nil) {
		resp.Err = rest.ErrorInternal("Internal error")
		return resp
	}

	if (modelFindEmail == nil) && (modelFindPhone == nil) {
		resp.Err = rest.ErrorNotFound("User not found")
		return resp
	}

	var UserIdTemp uint64
	if (modelFindEmail != nil) && (modelFindEmail.Id != 0) {
		UserIdTemp = modelFindEmail.Id

	}

	if (modelFindPhone != nil) && (modelFindPhone.Id != 0) {

		UserIdTemp = modelFindPhone.Id

	}
	modelAuthenticationHistory := &database.AuthenticationHistoryModel{

		UserId: UserIdTemp,
	}
	var err error
	if modelFindEmail.Id != 0 {
		modelAuthenticationHistory.SignOut = time.Now()
		err = s.db.AuthenticationHistory().Update(modelAuthenticationHistory)
	}
	if err != nil {
		s.log.Error(err)
		resp.Err = rest.ErrFieldResp{
			Meta: rest.ErrMeta{
				ErrCode:    500,
				ErrMessage: err.Error(),
			},
		}
		return resp
	}

	s.session.Delete(s.session.SessionId(ctx))

	return resp
}

func (s *service) RecoveryCodes(ctx context.Context, req RecoveryCodesRequest) (resp *RecoveryCodesResponse) {
	var modelMfa *database.MfaRecoveryCodesModel
	modelMfa = &database.MfaRecoveryCodesModel{
		UserId: req.UserId,
	}

	resp = &RecoveryCodesResponse{}

	codes, err := s.db.MfaRecoveryCodes().Create(modelMfa)
	if err != nil {
		s.log.Error(err)
		resp.Err = rest.ErrFieldResp{
			Meta: rest.ErrMeta{
				ErrCode:    500,
				ErrMessage: err.Error(),
			},
		}
		return resp
	}
	resp.Codes = codes
	return resp
}

func (s *service) SignInSocial() {

}

func (s *service) SignOutSocial(res http.ResponseWriter, req *http.Request) {
	if gothUser, err := gothic.CompleteUserAuth(res, req); err == nil {
		t, _ := template.New("User").Parse("userTemplate")
		t.Execute(res, gothUser)
	}
}

/*func (s *service) MakeProfileEndpoint(ctx context.Context,req ProfileRequest)(resp *ProfileRequest){
	return respe
}*/
var CompleteUserAuth = func(res http.ResponseWriter, req *http.Request) (goth.User, error) {
	defer Logout(res, req)
	if !keySet && defaultStore == Store {
		fmt.Println("goth/gothic: no SESSION_SECRET environment variable is set. The default cookie store is not available and any calls will fail. Ignore this warning if you are using a different store.")
	}

	providerName, err := GetProviderName(req)
	if err != nil {
		return goth.User{}, err
	}

	provider, err := goth.GetProvider(providerName)
	if err != nil {
		return goth.User{}, err
	}

	value, err := GetFromSession(providerName, req)
	if err != nil {
		return goth.User{}, err
	}

	sess, err := provider.UnmarshalSession(value)
	if err != nil {
		return goth.User{}, err
	}

	err = validateState(req, sess)
	if err != nil {
		return goth.User{}, err
	}

	user, err := provider.FetchUser(sess)
	if err == nil {
		// user can be found with existing session data
		return user, err
	}

	// get new token and retry fetch
	_, err = sess.Authorize(provider, req.URL.Query())
	if err != nil {
		return goth.User{}, err
	}

	err = StoreInSession(providerName, sess.Marshal(), req, res)

	if err != nil {
		return goth.User{}, err
	}

	gu, err := provider.FetchUser(sess)
	return gu, err
}
func Logout(res http.ResponseWriter, req *http.Request) error {
	session, err := Store.Get(req, SessionName)
	if err != nil {
		return err
	}
	session.Options.MaxAge = -1
	session.Values = make(map[interface{}]interface{})
	err = session.Save(req, res)
	if err != nil {
		return errors.New("Could not delete user session ")
	}
	return nil
}

func BeginAuthHandler(res http.ResponseWriter, req *http.Request) {
	url, err := GetAuthURL(res, req)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(res, err)
		return
	}

	http.Redirect(res, req, url, http.StatusTemporaryRedirect)
}