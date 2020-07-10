package controller

import (
	"bytes"
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
	template2 "html/template"
	"net/http"
	"net/url"
)

// Auth check the input password and username for authentication.
// ????username?password????user?role?permission?????menu
// ???????(goadmin_users)????(??)
func (h *Handler) Auth(ctx *context.Context) {

	var (
		user     models.UserModel
		ok       bool
		errMsg   = "fail"
		// ServiceKey = auth
		// ???????????(auth.ServiceKey)?Service
		s, exist = h.services.GetOrNot(auth.ServiceKey)
	)

	// ??Handler.captchaConfig(map[string]string)["driver"]??
	if capDriver, ok := h.captchaConfig["driver"]; ok {
		// Get?plugins\admin\modules\captcha\captcha.go
		// ??List(make(map[string]Captcha))??????key?????Captcha(interface)
		capt, ok := captcha.Get(capDriver)

		// ??token
		if ok {
			if !capt.Validate(ctx.FormValue("token")) {
				response.BadRequest(ctx, "wrong captcha")
				return
			}
		}
	}

	// ????key??url????
	if !exist {
		password := ctx.FormValue("password")
		username := ctx.FormValue("username")

		if password == "" || username == "" {
			response.BadRequest(ctx, "wrong password or username")
			return
		}
		// modules\auth\auth.go
		// ??user??????????user?role?permission???menu????????(goadmin_users)????(??)
		user, ok = auth.Check(password, username, h.conn)
	} else {
		user, ok, errMsg = auth.GetService(s).P(ctx)
	}

	if !ok {
		response.BadRequest(ctx, errMsg)
		return
	}

	// ??cookie(struct)????response header Set-Cookie?
	err := auth.SetCookie(ctx, user, h.conn)

	if err != nil {
		response.Error(ctx, err.Error())
		return
	}

	// ????Referer??Header
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

	// ?????code:200 and msg:ok and data
	response.OkWithData(ctx, map[string]interface{}{
		"url": h.config.GetIndexURL(),
	})
	return

}

// Logout delete the cookie.
func (h *Handler) Logout(ctx *context.Context) {
	// DelCookie??cookie(session)??
	// GetConnection?????service.List?????Connection(interface)??
	err := auth.DelCookie(ctx, db.GetConnection(h.services))
	if err != nil {
		logger.Error("logout error", err)
	}

	// AddHeader???(key?value)??header?(Context.Response.Header)
	// GetLoginUrl globalCfg.LoginUrl
	ctx.AddHeader("Location", h.config.Url(config.GetLoginUrl()))
	ctx.SetStatusCode(302)
}

// ShowLogin show the login page.
// ShowLogin??map[string]Component(interface)?????login(key)???????template?data??buf???HTML
func (h *Handler) ShowLogin(ctx *context.Context) {

	// GetComp??map[string]Component?????name(login)?????????Component(interface)
	// GetTemplate?Component(interface)???
	tmpl, name := template.GetComp("login").GetTemplate()
	buf := new(bytes.Buffer)

	// ExecuteTemplate?html/template??
	// ??????data??buf(struct)???HTML
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
