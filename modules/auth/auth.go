// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package auth

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/service"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules"
	"golang.org/x/crypto/bcrypt"
	"sync"
)

// Auth get the user model from Context.
// 透過參數ctx回傳目前登入的用戶(Context.UserValue["user"])並轉換成UserModel
func Auth(ctx *context.Context) models.UserModel {
	// User回傳目前登入的用戶(Context.UserValue["user"])並轉換成UserModel
	return ctx.User().(models.UserModel)
}

// Check check the password and username and return the user model.
// 檢查user密碼是否正確之後取得user的role、permission及可用menu，最後更新資料表(goadmin_users)的密碼值(加密)
func Check(password string, username string, conn db.Connection) (user models.UserModel, ok bool) {

	// plugins\admin\models\user.go
	// User設置UserModel.Base.TableName(struct)並回傳設置UserModel(struct)
	// SetConn將參數conn(db.Connection)設置至UserModel.conn(UserModel.Base.Conn)
	user = models.User().SetConn(conn).FindByUserName(username)

	// 判斷user是否為空
	if user.IsEmpty() {
		ok = false
	} else {
		// 檢查密碼
		if comparePassword(password, user.Password) {
			ok = true
			//取得user的role、permission及可用menu
			user = user.WithRoles().WithPermissions().WithMenus()
			// EncodePassword將參數pwd加密
			// UpdatePwd將參數設置至UserModel.UserModel並且更新dialect.H{"password": password,}
			user.UpdatePwd(EncodePassword([]byte(password)))
		} else {
			ok = false
		}
	}
	return
}

// 檢查密碼是否相符
func comparePassword(comPwd, pwdHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(pwdHash), []byte(comPwd))
	return err == nil
}

// EncodePassword encode the password.
// 將參數pwd加密
func EncodePassword(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hash[:])
}

// SetCookie set the cookie.
// 設置cookie(struct)並儲存在response header Set-Cookie中
func SetCookie(ctx *context.Context, user models.UserModel, conn db.Connection) error {
	// 設置Session(struct)資訊並取得cookie及設置cookie值
	ses, err := InitSession(ctx, conn)

	if err != nil {
		return err
	}

	// Add將參數"user_id"、user.Id加入Session.Values後檢查是否有符合Session.Sid的資料，判斷插入或是更新資料
	// 最後設置cookie(struct)並儲存在response header Set-Cookie中
	return ses.Add("user_id", user.Id)
}

// DelCookie delete the cookie from Context.
// 清除cookie(session)資料
func DelCookie(ctx *context.Context, conn db.Connection) error {
	// 設置Session(struct)資訊並取得cookie及設置cookie值
	ses, err := InitSession(ctx, conn)

	if err != nil {
		return err
	}

	// 清除cookie(session)
	return ses.Clear()
}

type TokenService struct {
	tokens CSRFToken //[]string
	lock   sync.Mutex
}

// 回傳"token_csrf_helper"
func (s *TokenService) Name() string {
	// TokenServiceKey = token_csrf_helper
	return TokenServiceKey
}

func init() {
	// TokenServiceKey = token_csrf_helper
	// Register將參數TokenServiceKey、gen(函式)將入services(map[string]Generator)中
	service.Register(TokenServiceKey, func() (service.Service, error) {
		// 設置並回傳TokenService.tokens(struct)
		return &TokenService{
			tokens: make(CSRFToken, 0),
		}, nil
	})
}

const (
	TokenServiceKey = "token_csrf_helper"
	ServiceKey      = "auth"
)

// 將參數s轉換成TokenService(struct)類別後回傳
func GetTokenService(s interface{}) *TokenService {
	if srv, ok := s.(*TokenService); ok {
		return srv
	}
	panic("wrong service")
}

// AddToken add the token to the CSRFToken.
// 建立uuid並設置至TokenService.tokens，回傳uuid(string)
func (s *TokenService) AddToken() string {
	s.lock.Lock()
	defer s.lock.Unlock()
	// 設置uuid
	tokenStr := modules.Uuid()
	s.tokens = append(s.tokens, tokenStr)
	return tokenStr
}

// CheckToken check the given token with tokens in the CSRFToken, if exist
// return true.
// 檢查TokenService.tokens([]string)裡是否有符合參數toCheckToken的值
// 如果符合，將在TokenService.tokens([]string)裡將符合的toCheckToken從[]string拿出
func (s *TokenService) CheckToken(toCheckToken string) bool {
	for i := 0; i < len(s.tokens); i++ {
		if (s.tokens)[i] == toCheckToken {
			s.tokens = append((s.tokens)[:i], (s.tokens)[i+1:]...)
			return true
		}
	}
	return false
}

// CSRFToken is type of a csrf token list.
type CSRFToken []string

type Processor func(ctx *context.Context) (model models.UserModel, exist bool, msg string)

type Service struct {
	P Processor
}

// 回傳auth(string)
func (s *Service) Name() string {
	return "auth"
}

// 將參數s傳換成Service(struct)後回傳
func GetService(s interface{}) *Service {
	if srv, ok := s.(*Service); ok {
		return srv
	}
	panic("wrong service")
}

// 將參數processor設置至Service.P(struct)並回傳
func NewService(processor Processor) *Service {
	return &Service{
		P: processor,
	}
}
