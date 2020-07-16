package controller

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/guard"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/response"
)

// Delete delete the row from database.
// 透過id刪除資料後回傳code、data(token)、msg
func (h *Handler) Delete(ctx *context.Context) {

	// 取得Context.UserValue[delete_param]的值並轉換成DeleteParam(struct)
	param := guard.GetDeleteParam(ctx)

	//token := ctx.FormValue("_t")
	//
	//if !auth.TokenHelper.CheckToken(token) {
	//	ctx.SetStatusCode(http.StatusBadRequest)
	//	ctx.WriteString(`{"code":400, "msg":"delete fail"}`)
	//	return
	//}

	// 透過id刪除資料
	// param.Prefix = manager、roles、permission
	if err := h.table(param.Prefix, ctx).DeleteData(param.Id); err != nil {
		logger.Error(err)
		response.Error(ctx, "delete fail")
		return
	}
	// authSrv將參數h.services.Get(auth.TokenServiceKey)轉換成TokenService(struct)類別後回傳
	// AddToken建立uuid並設置至TokenService.tokens，回傳uuid(string)
	response.OkWithData(ctx, map[string]interface{}{
		"token": h.authSrv().AddToken(),
	})
}
