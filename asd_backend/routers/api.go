/**
 * api-路由
 * @author Evotrek 研发团队
 * @since 2025-1-27
 * @File : api
 */
package routers

import (
	"asd/app/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	beego.Router("/api/login/send-code", &controllers.UserApiController{}, "post:SendSmsCode")
	beego.Router("/api/login/mobile", &controllers.UserApiController{}, "post:MobileLogin")
	beego.Router("/api/login/wechat", &controllers.UserApiController{}, "get:WechatLoginUrl")
	beego.Router("/api/login/wechat", &controllers.UserApiController{}, "post:WechatLoginParams")
	beego.Router("/api/login/wechat/callback", &controllers.UserApiController{}, "get,post:WechatLoginCallback")
	beego.Router("/api/login/mobile-pwd", &controllers.UserApiController{}, "post:MobilePwdLogin")

	beego.Router("/api/users/profile", &controllers.UserApiController{}, "post:UserProfile")
	beego.Router("/api/logout", &controllers.UserApiController{}, "post:Logout")

	beego.Router("/api/chat/list", &controllers.ChatApiController{}, "get,post:List")
	beego.Router("/api/chat/update", &controllers.ChatApiController{}, "post:Update")
	beego.Router("/api/chat/delete", &controllers.ChatApiController{}, "post:Delete")
	beego.Router("/api/chat/detail", &controllers.ChatApiController{}, "post:Detail")

	beego.Router("/api/chat/stop_stream", &controllers.ChatApiController{}, "post:StopStream")
	beego.Router("/api/chat/create_stream", &controllers.ChatApiController{}, "post:CreateStream")
	beego.Router("/api/chat/read_stream", &controllers.ChatApiController{}, "post:ReadStream")

	beego.Router("/api/files/upload", &controllers.FileApiController{}, "post:Upload")
	beego.Router("/api/files/detail", &controllers.FileApiController{}, "post:Get")

	beego.Router("/api/reports/list", &controllers.ReportApiController{}, "get,post:List")

	beego.Router("/api/admin/report/upload", &controllers.AdminApiController{}, "post:AddReport")
	beego.Router("/api/admin/messages/fix-all", &controllers.AdminApiController{}, "post:FixPubMedLinks")
	beego.Router("/api/admin/rag/status", &controllers.AdminApiController{}, "post:GetRagWorkersStatus")

}
