package controller

import (
	"bytes"
	template2 "html/template"
	"net/http"
	"net/url"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/modules/system"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/captcha"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/response"
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/types"
)

// Auth check the input password and username for authentication.
// 身分驗證username、password，接著取得user、role、permission及可用menu
// 最後更新資料表(goadmin_users)的密碼值(加密)
func (h *Handler) Auth(ctx *context.Context) {

	var (
		user   models.UserModel
		ok     bool
		errMsg = "fail"
		// ServiceKey = auth
		// 判斷是否有取得符合參數(auth.ServiceKey)的Service
		s, exist = h.services.GetOrNot(auth.ServiceKey)
	)

	// 取得Handler.captchaConfig(map[string]string)["driver"]的值
	if capDriver, ok := h.captchaConfig["driver"]; ok {
		// Get在plugins\admin\modules\captcha\captcha.go
		// 判斷List(make(map[string]Captcha))裡是否有參數key的值並回傳Captcha(interface)
		capt, ok := captcha.Get(capDriver)

		// 驗證token
		if ok {
			if !capt.Validate(ctx.FormValue("token")) {
				response.BadRequest(ctx, "wrong captcha")
				return
			}
		}
	}

	// 藉由參數key取得url的參數值
	if !exist {
		password := ctx.FormValue("password")
		username := ctx.FormValue("username")

		if password == "" || username == "" {
			response.BadRequest(ctx, "wrong password or username")
			return
		}
		// modules\auth\auth.go
		// 檢查user密碼是否正確之後取得user的role、permission及可用menu，最後更新資料表(goadmin_users)的密碼值(加密)
		user, ok = auth.Check(password, username, h.conn)
	} else {
		user, ok, errMsg = auth.GetService(s).P(ctx)
	}

	if !ok {
		response.BadRequest(ctx, errMsg)
		return
	}

	// 設置cookie(struct)並儲存在response header Set-Cookie中
	err := auth.SetCookie(ctx, user, h.conn)

	if err != nil {
		response.Error(ctx, err.Error())
		return
	}

	// 藉由參數Referer獲得Header
	if ref := ctx.Headers("Referer"); ref != "" {
		if u, err := url.Parse(ref); err == nil {
			v := u.Query()
			if r := v.Get("ref"); r != "" {
				rr, _ := url.QueryUnescape(r)
				response.OkWithData(ctx, map[string]interface{}{
					"url": rr,
				})
				return
			}
		}
	}

	// 成功，回傳code:200 and msg:ok and data
	response.OkWithData(ctx, map[string]interface{}{
		"url": h.config.GetIndexURL(),
	})
	return

}

// Logout delete the cookie.
func (h *Handler) Logout(ctx *context.Context) {
	// DelCookie清除cookie(session)資料
	// GetConnection取得匹配的service.List然後轉換成Connection(interface)類別
	err := auth.DelCookie(ctx, db.GetConnection(h.services))
	if err != nil {
		logger.Error("logout error", err)
	}

	// AddHeader將參數(key、value)添加header中(Context.Response.Header)
	// GetLoginUrl globalCfg.LoginUrl
	ctx.AddHeader("Location", h.config.Url(config.GetLoginUrl()))
	ctx.SetStatusCode(302)
}

// ShowLogin show the login page.
// ShowLogin判斷map[string]Component(interface)是否有參數login(key)的值，接著執行template將data寫入buf並輸出HTML
func (h *Handler) ShowLogin(ctx *context.Context) {

	// GetComp判斷map[string]Component是否有參數name(login)的值，有的話則回傳Component(interface)
	// GetTemplate為Component(interface)的方法
	tmpl, name := template.GetComp("login").GetTemplate()
	buf := new(bytes.Buffer)

	// ExecuteTemplate為html/template套件
	// 將第三個參數data寫入buf(struct)後輸出HTML
	if err := tmpl.ExecuteTemplate(buf, name, struct {
		UrlPrefix string
		Title     string
		Logo      template2.HTML
		CdnUrl    string
		System    types.SystemInfo
	}{
		UrlPrefix: h.config.AssertPrefix(),
		Title:     h.config.LoginTitle,
		Logo:      h.config.LoginLogo,
		System: types.SystemInfo{
			Version: system.Version(),
		},
		CdnUrl: h.config.AssetUrl,
	}); err == nil {
		ctx.HTML(http.StatusOK, buf.String())
	} else {
		logger.Error(err)
		ctx.HTML(http.StatusOK, "parse template error (；′⌒`)")
	}
}
