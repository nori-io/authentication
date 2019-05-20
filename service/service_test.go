package service_test

import (
	"context"
	"database/sql/driver"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	rest "github.com/cheebo/gorest"
	"github.com/nori-io/nori-common/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/nori-io/authentication/service"
	"github.com/nori-io/authentication/service/database"
)

type AnyTime struct {
}

type AnyByteArray struct {
}

func TestService_SignUp_Email_UserExists(t *testing.T) {
	auth := &mocks.Auth{}

	cache := &mocks.Cache{}
	cfg := &service.Config{}
	mockDatabase, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	db := database.DB(mockDatabase, logrus.New())
	mail := &mocks.Mail{}
	session := &mocks.Session{}

	serviceTest := service.NewService(auth, cache, cfg, db, new(logrus.Logger), mail, session)
	signUpRequest := service.SignUpRequest{Email: "test@mail.ru", Password: "pass"}
	errField := rest.ErrFieldResp{
		Meta: rest.ErrFieldRespMeta{
			ErrCode:    0,
			ErrMessage: "",
		},
	}
	errField.AddError("phone, email", 400, "User already exists.")

	respExpected := service.SignUpResponse{Id: 0, Email: "test@mail.ru", PhoneNumber: "", PhoneCountryCode: "", Err: errField}

	nonEmptyRows := sqlmock.NewRows([]string{"id", "email", "password", "salt"}).
		AddRow(1, "test@mail.ru", "pass", "salt")

	mock.ExpectQuery("SELECT id, email,password,salt FROM auth WHERE email = ? LIMIT 1").
		WithArgs("test@mail.ru").WillReturnRows(nonEmptyRows)

	//mock.ExpectQuery("SELECT id, email,password,salt FROM auth WHERE email = ? LIMIT 1").WillReturnRows(nil)

	resp := serviceTest.SignUp(context.Background(), signUpRequest)

	assert.Equal(t, &respExpected, resp)
}

func TestService_SignUp_Email_UserNotExist(t *testing.T) {
	auth := &mocks.Auth{}

	cache := &mocks.Cache{}
	cfg := &service.Config{}
	mockDatabase, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	db := database.DB(mockDatabase, logrus.New())
	mail := &mocks.Mail{}
	session := &mocks.Session{}

	serviceTest := service.NewService(auth, cache, cfg, db, new(logrus.Logger), mail, session)
	signUpRequest := service.SignUpRequest{Email: "test@mail.ru", Password: "pass"}

	respExpected := service.SignUpResponse{Id: 0, Email: "test@mail.ru", PhoneNumber: "", PhoneCountryCode: "", Err: nil}

	/*	nonEmptyRows := sqlmock.NewRows([]string{"id", "email", "password", "salt"}).
			AddRow(1, "test@mail.ru", "pass", "salt")

		mock.ExpectQuery("SELECT id, email,password,salt FROM auth WHERE email = ? LIMIT 1").
			WithArgs("test@mail.ru").WillReturnRows(nonEmptyRows)*/

	mock.ExpectQuery("SELECT id, email,password,salt FROM auth WHERE email = ? LIMIT 1").WillReturnRows(nil)

	mock.ExpectBegin()
	mock.ExpectPrepare("INSERT INTO users (status_account, type, created, updated) VALUES(?,?,?,?)").
		ExpectExec().WithArgs("locked", "", AnyTime{}, AnyTime{}).
		WillReturnResult(sqlmock.NewResult(1, 1))

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(1).
		RowError(1, fmt.Errorf("row error"))
	mock.ExpectQuery("SELECT LAST_INSERT_ID()").WillReturnRows(rows)

	mock.ExpectPrepare("INSERT INTO auth (user_id, email, password, salt, created, updated, is_email_verified, is_phone_verified) VALUES(?,?,?,?,?,?,?,?)").
		ExpectExec().
		WithArgs(1, "test@mail.ru", AnyByteArray{}, AnyByteArray{}, AnyTime{}, AnyTime{}, false, false).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	resp := serviceTest.SignUp(context.Background(), signUpRequest)

	assert.Equal(t, &respExpected, resp)
}

func TestService_SignUp_Phone_UserExists(t *testing.T) {
	auth := &mocks.Auth{}

	cache := &mocks.Cache{}
	cfg := &service.Config{}
	mockDatabase, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	db := database.DB(mockDatabase, logrus.New())
	mail := &mocks.Mail{}
	session := &mocks.Session{}

	serviceTest := service.NewService(auth, cache, cfg, db, new(logrus.Logger), mail, session)
	errField := rest.ErrFieldResp{
		Meta: rest.ErrFieldRespMeta{
			ErrCode:    0,
			ErrMessage: "",
		},
	}
	errField.AddError("phone, email", 400, "User already exists.")

	signUpRequest := service.SignUpRequest{PhoneCountryCode: "1", PhoneNumber: "234567890", Password: "pass"}

	//respExcepted.Err = errField

	respExpected := service.SignUpResponse{Id: 0, Email: "", PhoneCountryCode: "1", PhoneNumber: "234567890", Err: errField}

	nonEmptyRows := sqlmock.NewRows([]string{"id", "phone_country_code", "phone_number", "password", "salt"}).
		AddRow(1, "1", "234567890", "pass", "salt")

	mock.ExpectQuery("SELECT id, phone_country_code, phone_number, password,salt FROM auth WHERE concat(phone_country_code,phone_number)=?  LIMIT 1").WithArgs("1234567890").WillReturnRows(nonEmptyRows)

	mock.ExpectBegin()
	db.Users().Create(&database.AuthModel{
		Email:    "test@mail.ru",
		Password: []byte("pass"),
		Salt:     []byte("salt"),
	}, &database.UsersModel{
		Status_account: "active",
		Type:           "vendor",
	})

	//mock.ExpectQuery("SELECT id, email,password,salt FROM auth WHERE email = ? LIMIT 1").WillReturnRows(nil)

	resp := serviceTest.SignUp(context.Background(), signUpRequest)

	assert.Equal(t, &respExpected, resp)
}

/*func TestService_ActivationCode(t *testing.T) {

	auth := &mocks.Auth{}

	cache := &mocks.Cache{}
	cfg := &service.Config{}
	mockDatabase, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	db := database.DB(mockDatabase, logrus.New())
	mail := &mocks.Mail{}
	session := &mocks.Session{}

	//mail.On("Send").

	//serviceTest := service.NewService(auth, cache, cfg, db, new(logrus.Logger), mail, session)

	//SignUpRequest:=service.SignUpRequest{Email:"test@mail.ru", Password:"pass", Type:"vendor"}

	//serviceTest.SignUp(context.Background(), SignUpRequest)

	code := rand.RandomAlphaNum(65)
	salt, codeForSend, err := getActivationCode([]byte(code))
	if err == nil {
		//activationRequest:=service.ActivationCodeRequest{ActivationCode:codeForSend}
		activate([]byte(code), salt, codeForSend)
		mail.Send(activate([]byte(code), salt, codeForSend))
	}

}

func getActivationCode(codeMessage []byte) ([]byte, []byte, error) {
	salt := rand.RandomAlphaNum(65)

	codeForSend, err := database.Hash(codeMessage, []byte(salt))

	return []byte(salt), codeForSend, err
}

func activate(code []byte, salt []byte, codeForSend []byte) string {
	result, _ := database.VerifyPassword([]byte(code), salt, codeForSend)
	b := strconv.FormatBool(result)

	return b
}*/

func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func (s AnyByteArray) Match(v driver.Value) bool {
	_, ok := v.([]byte)
	return ok
}
