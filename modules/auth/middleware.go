// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package auth

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/constant"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/errors"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/modules/page"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	template2 "github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/types"
	"net/http"
	"net/url"
)

// Invoker contains the callback functions which are used
// in the route middleware.
type Invoker struct {
	prefix                 string
	// MiddlewareCallback的類別為func(ctx *context.Context)
	authFailCallback       MiddlewareCallback //驗證失敗Callback
	permissionDenyCallback MiddlewareCallback //權限拒絕Callback
	conn                   db.Connection
}

// Middleware is the default auth middleware of plugins.
// 建立Invoker(Struct)並透過參數ctx取得UserModel，並且取得該user的role、權限與可用menu，最後檢查用戶權限
func Middleware(conn db.Connection) context.Handler {
	// DefaultInvoker設置並回傳Invoker(Struct)
	// Middleware透過參數ctx取得UserModel，並且取得該user的role、權限與可用menu，最後檢查用戶權限
	// 如驗證及權限通過，將參數"user"、user設置至Context.UserValue(map[string]interface{})並執行迴圈的func(ctx *Context)
	// 若失敗，則調用Invoker的authFailCallback或permissionDenyCallback(function)
	return DefaultInvoker(conn).Middleware()
}

// DefaultInvoker return a default Invoker.
// 設置並回傳Invoker(Struct)
func DefaultInvoker(conn db.Connection) *Invoker {
	return &Invoker{
		// 回傳globalCfg.prefix
		prefix: config.Prefix(),
		authFailCallback: func(ctx *context.Context) {
			if ctx.Request.URL.Path == config.Url(config.GetLoginUrl()) {
				return
			}
			if ctx.Request.URL.Path == config.Url("/logout") {
				// Write將狀態碼，標頭及body寫入Context.Response
				ctx.Write(302, map[string]string{
					"Location": config.Url(config.GetLoginUrl()),
				}, ``)
				return
			}
			param := ""
			// url後加入參數
			if ref := ctx.Headers("Referer"); ref != "" {
				param = "?ref=" + url.QueryEscape(ref)
			}

			u := config.Url(config.GetLoginUrl() + param)
			_, err := ctx.Request.Cookie(DefaultCookieKey)
			referer := ctx.Headers("Referer")

			// constant.PjaxHeade = X-PJAX
			if (ctx.Headers(constant.PjaxHeader) == "" && ctx.Method() != "GET") ||
				err != nil ||
				referer == "" {
				ctx.Write(302, map[string]string{
					"Location": u,
				}, ``)
			} else {
				// 登入時間過長或同個IP登入
				msg := language.Get("login overdue, please login again")
				//添加HTML
				ctx.HTML(http.StatusOK, `<script>
	if (typeof(swal) === "function") {
		swal({
			type: "info",
			title: `+language.Get("login info")+`,
			text: "`+msg+`",
			showCancelButton: false,
			confirmButtonColor: "#3c8dbc",
			confirmButtonText: '`+language.Get("got it")+`',
        })
		setTimeout(function(){ location.href = "`+u+`"; }, 3000);
	} else {
		alert("`+msg+`")
		location.href = "`+u+`"
    }
</script>`)
			}
		},
		permissionDenyCallback: func(ctx *context.Context) {
			// constant.PjaxHeade = X-PJAX
			if ctx.Headers(constant.PjaxHeader) == "" && ctx.Method() != "GET" {
				// 轉換成JSON存至Context.Response.body
				ctx.JSON(http.StatusForbidden, map[string]interface{}{
					"code": http.StatusForbidden,
					"msg":  language.Get(errors.PermissionDenied),
				})
			} else {
				page.SetPageContent(ctx, Auth(ctx), func(ctx interface{}) (types.Panel, error) {
					return template2.WarningPanel(errors.PermissionDenied, template2.NoPermission403Page), nil
				}, conn)
			}
		},
		conn: conn,
	}
}

// SetPrefix return the default Invoker with the given prefix.
// 透過conn參數建立Invoker並將參數prefix設置至Invoker.prefix
func SetPrefix(prefix string, conn db.Connection) *Invoker {
	// DefaultInvoker設置並回傳Invoker(Struct)
	i := DefaultInvoker(conn)
	i.prefix = prefix
	return i
}

// SetAuthFailCallback set the authFailCallback of Invoker.
// 將參數callback設置至Invoker.authFailCallback(struct)
func (invoker *Invoker) SetAuthFailCallback(callback MiddlewareCallback) *Invoker {
	invoker.authFailCallback = callback
	return invoker
}

// SetPermissionDenyCallback set the permissionDenyCallback of Invoker.
// 將參數callback設置至Invoker.permissionDenyCallback(struct)
func (invoker *Invoker) SetPermissionDenyCallback(callback MiddlewareCallback) *Invoker {
	invoker.permissionDenyCallback = callback
	return invoker
}

// MiddlewareCallback is type of callback function.
type MiddlewareCallback func(ctx *context.Context)

// Middleware get the auth middleware from Invoker.
// 透過參數ctx取得UserModel，並且取得該user的role、權限與可用menu，最後檢查用戶權限
// 如驗證及權限通過，將參數"user"、user設置至Context.UserValue(map[string]interface{})並執行迴圈的func(ctx *Context)
// 若失敗，則調用Invoker的authFailCallback或permissionDenyCallback(function)
func (invoker *Invoker) Middleware() context.Handler {
	return func(ctx *context.Context) {

		// 透過參數ctx取得UserModel，並且取得該user的role、權限與可用menu，最後檢查用戶權限
		user, authOk, permissionOk := Filter(ctx, invoker.conn)

		if authOk && permissionOk {
			// 將參數"user"、user設置至Context.UserValue(map[string]interface{})
			ctx.SetUserValue("user", user)
			// 執行迴圈的func(ctx *Context)
			ctx.Next()
			return
		}
		// 驗證及權限通過
		if !authOk {
			invoker.authFailCallback(ctx)
			ctx.Abort()
			return
		}

		if !permissionOk {
			ctx.SetUserValue("user", user)
			// 執行Invoker.authFailCallback(驗證失敗function)
			invoker.permissionDenyCallback(ctx)
			// 設置Context.index = 63
			ctx.Abort()
			return
		}
	}
}

// Filter retrieve the user model from Context and check the permission
// at the same time.
// 透過參數ctx取得UserModel，並且取得該user的role、權限與可用menu，最後檢查用戶權限
func Filter(ctx *context.Context, conn db.Connection) (models.UserModel, bool, bool) {
	var (
		id   float64
		ok   bool
		// 取得預設的UserModel
		user = models.User()
	)

	// 設置Session(struct)資訊並取得cookie及設置cookie值
	ses, err := InitSession(ctx, conn)

	if err != nil {
		// 驗證失敗
		logger.Error("retrieve auth user failed", err)
		return user, false, false
	}

	// 藉由參數取得Session.Values[user_id]
	if id, ok = ses.Get("user_id").(float64); !ok {
		return user, false, false
	}

	// GetCurUserByID取得該id的角色、權限以即可訪問的菜單
	user, ok = GetCurUserByID(int64(id), conn)

	if !ok {
		return user, false, false
	}

	// CheckPermissions透過path、method、param檢查用戶權限
	return user, true, CheckPermissions(user, ctx.Request.URL.String(), ctx.Method(), ctx.PostForm())
}

const defaultUserIDSesKey = "user_id"


// GetUserID return the user id from the session.
// 尋找資料表中符合參數(sesKey)的user資料，將資料的values欄位值JSON解碼並回傳values(map)["user_id"]鍵的值(id)
func GetUserID(sesKey string, conn db.Connection) int64 {
	// 在/modules/auth/session.go中
	// defaultUserIDSesKey = "user_id"
	// GetSessionByKey尋找資料表中符合參數(sesKey)的user資料，將資料的values欄位值JSON解碼並回傳values(map)["user_id"]鍵的值(id)
	id, err := GetSessionByKey(sesKey, defaultUserIDSesKey, conn)
	if err != nil {
		logger.Error("retrieve auth user failed", err)
		return -1
	}
	if idFloat64, ok := id.(float64); ok {
		return int64(idFloat64)
	}
	return -1
}

// GetCurUser return the user model.
// 透過參數sesKey(cookie)取得id並利用id取得該user的role、permission以及可用menu並回傳UserModel(struct)
func GetCurUser(sesKey string, conn db.Connection) (user models.UserModel, ok bool) {

	if sesKey == "" {
		ok = false
		return
	}

	// 取得user_id(在goadmin_session資料表values欄位)
	// 尋找資料表中符合參數(sesKey)的user資料，將資料的values欄位值JSON解碼並回傳values(map)["user_id"]鍵的值(id)
	id := GetUserID(sesKey, conn)
	if id == -1 {
		ok = false
		return
	}

	// GetCurUserByID取得參數id的role、permission以及可使用menu並回傳UserModel(struct)
	return GetCurUserByID(id, conn)
}

// GetCurUserByID return the user model of given user id.
// 透過user_id尋找符合的UserModel
// models.UserModel在plugins/admin/models/user.go中
// 取得參數id的role、permission以及可使用menu並回傳UserModel(struct)
func GetCurUserByID(id int64, conn db.Connection) (user models.UserModel, ok bool) {

	// models.User() 在plugins/admin/models/user.go中
	// models.User() 回傳預設UserModel
	// SetConn 將參數conn設置至UserModel.conn(UserModel.Base.Conn)
	// Find透過id尋找符合的UserModel
	user = models.User().SetConn(conn).Find(id)

	if user.IsEmpty() {
		ok = false
		return
	}

	//判斷是否有頭像
	if user.Avatar == "" || config.GetStore().Prefix == "" {
		user.Avatar = ""
	} else {
		//GetStore、URL 在congig/cinfig.go中，取得儲存路徑
		user.Avatar = config.GetStore().URL(user.Avatar)
	}

	// WithRoles、WithPermissions、WithMenus都在plugins/admin/models/user.go
	// 取得角色、權限及可使用菜單
	user = user.WithRoles().WithPermissions().WithMenus()

	// HasMenu在plugins/admin/models/user.go
	// 檢查用戶是否有可訪問的menu
	ok = user.HasMenu()

	return
}

// CheckPermissions check the permission of the user.
// CheckPermissions透過path、method、param檢查用戶權限
func CheckPermissions(user models.UserModel, path, method string, param url.Values) bool {
	// CheckPermissionByUrlMethod在plugins\admin\models\user.go中
	return user.CheckPermissionByUrlMethod(path, method, param)
}
