package controllers

import (
	"asd/app/dto"
	"asd/app/models"
	"asd/app/vo"
	"asd/conf"
	"asd/utils"
	"asd/utils/beego"
	"asd/utils/common"
	"asd/utils/sms"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/beego/beego/v2/core/logs"
)

var UserApi = new(UserApiController)

type UserApiController struct {
	BaseController
}

var (
	DEFAULT_USER_NICKNAME = "星星"
	DEFAULT_USER_AVATAR   = "https://public-1306205170.cos.ap-nanjing.myqcloud.com/images/avatar.jpg"
	DEFAULT_USER_PASSWORD = "123456"
)

// generateState 生成安全的随机字符串作为 state
func generateState() (string, error) {
	b := make([]byte, 16) // 16 bytes -> 32 hex characters
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (c *UserApiController) SendSmsCode() {
	var req dto.SendSmsCodeReq
	if err := c.ParseJSON(&req); err != nil {
		c.ErrorJson(400, "请求参数错误"+err.Error())
		return
	}

	if !sms.IsValidMobile(req.Mobile) {
		c.ErrorJson(400, "无效的手机号码")
		return
	}

	// Generate a random 6-digit code
	code := fmt.Sprintf("%06d", common.RandomInt(100000, 999999))

	if err := sms.SendSmsCode(req.Mobile, code); err != nil {
		c.ErrorJson(500, "发送短信失败: "+err.Error())
		return
	}

	c.JSON(common.JsonResult{
		Code: 0,
		Msg:  "验证码已发送",
	})
}

func (c *UserApiController) MobileLogin() {
	var req dto.UserMobileLoginReq
	if err := c.ParseJSON(&req); err != nil {
		c.ErrorJson(400, "请求参数错误"+err.Error())
		return
	}

	// validate the passcode from sms
	if !sms.VerifySmsCode(req.Mobile, req.Passcode) {
		c.ErrorJson(400, "验证码错误")
		return
	}

	user := models.User{
		Mobile: sql.NullString{String: req.Mobile, Valid: true},
	}

	if err := user.Get(); err == nil {
		// user exists, nothing to do
	} else {
		// new user, register
		user.Nickname = DEFAULT_USER_NICKNAME
		user.Avatar = DEFAULT_USER_AVATAR
		id, err := user.Insert()
		if err != nil {
			c.ErrorJson(http.StatusNotFound, err.Error())
			return
		}
		user.Id = int(id)
	}

	c.SetSession(conf.USER_ID, user.Id)

	logs.Info("User logged in successfully via Mobile.")
	c.JSON(common.JsonResult{
		Count: 1,
		Data: vo.UserInfoVo{
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
		},
	})
}

// Login initiates the WeChat login process
// GET /api/login/wechat
func (c *UserApiController) WechatLoginUrl() {
	appID := conf.CONFIG.TencentConfig.WechatAppID
	redirectURI := conf.CONFIG.TencentConfig.WechatRedirectURI

	if appID == "" || redirectURI == "" {
		c.ErrorJson(500, "WeChat AppID or RedirectURI not configured in app.conf")
		return
	}

	// 1. 生成 state 并存储到 Session
	state, err := generateState()
	if err != nil {
		c.ErrorJson(500, fmt.Sprintf("Failed to generate state for WeChat login: %v", err))
		return
	}
	// 使用 Beego 的 Session 存储 state
	c.SetSession(conf.WECHAT_LOGIN_STATE, state)

	// 2. 构造微信授权 URL
	authURL := fmt.Sprintf("https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect",
		appID,
		url.QueryEscape(redirectURI), // redirect_uri 需要 URL Encode
		state,
	)

	// 3. 重定向用户到微信授权页
	c.Redirect(authURL, http.StatusFound)
}

// Login initiates the WeChat login process
// POST /api/login/wechat
func (c *UserApiController) WechatLoginParams() {
	appID := conf.CONFIG.TencentConfig.WechatAppID
	redirectURI := conf.CONFIG.TencentConfig.WechatRedirectURI

	if appID == "" || redirectURI == "" {
		c.ErrorJson(500, "WeChat AppID or RedirectURI not configured in app.conf")
		return
	}

	// 1. 生成 state 并存储到 Session
	state, err := generateState()
	if err != nil {
		c.ErrorJson(500, fmt.Sprintf("Failed to generate state for WeChat login: %v", err))
		return
	}
	// 使用 Beego 的 Session 存储 state
	c.SetSession(conf.WECHAT_LOGIN_STATE, state)

	// 2. 重定向用户到微信授权页
	c.JSON(common.JsonResult{
		Code: 0,
		Data: beego.H{
			"appid":        appID,
			"scope":        "snsapi_login",
			"redirect_uri": redirectURI,
			"state":        state,
		},
	})
}

func (c *UserApiController) WechatLoginCallback() {
	appID := conf.CONFIG.TencentConfig.WechatAppID
	appSecret := conf.CONFIG.TencentConfig.WechatAppSecret
	successRedirect := "/"
	failRedirect := "/"

	// 1. 获取 code 和 state 参数
	code := c.GetString("code")
	receivedState := c.GetString("state")

	if code == "" {
		logs.Warn("WeChat callback missing 'code' parameter.")
		c.Redirect(failRedirect, http.StatusFound)
		return
	}

	// 2. 验证 state (CSRF 防护)
	storedState := c.GetSession(conf.WECHAT_LOGIN_STATE)
	if storedState == nil {
		logs.Warn("WeChat callback missing state in session.")
		c.Redirect(failRedirect, http.StatusFound)
		return
	}
	// 删除 session 中的 state, 防止重放攻击
	c.DelSession(conf.WECHAT_LOGIN_STATE)

	if receivedState == "" || receivedState != storedState.(string) {
		logs.Warn("WeChat callback state mismatch. Received: %s, Expected: %s", receivedState, storedState.(string))
		c.Redirect(failRedirect, http.StatusFound)
		return
	}

	// 3. 使用 code 换取 access_token 和 openid
	tokenURL := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		appID,
		appSecret,
		code,
	)

	resp, err := http.Get(tokenURL)
	if err != nil {
		logs.Error("Failed to request WeChat access token: %v", err)
		c.Redirect(failRedirect, http.StatusFound)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.Error("Failed to read WeChat access token response body: %v", err)
		c.Redirect(failRedirect, http.StatusFound)
		return
	}

	// 用于解析微信接口返回的 JSON
	type WechatAccessTokenResponse struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int64  `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		OpenID       string `json:"openid"`
		Scope        string `json:"scope"`
		UnionID      string `json:"unionid"` // 可能为空
		ErrCode      int    `json:"errcode"` // 获取失败时返回
		ErrMsg       string `json:"errmsg"`  // 获取失败时返回
	}

	var tokenResp WechatAccessTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		logs.Error("Failed to unmarshal WeChat access token response: %v. Body: %s", err, string(body))
		c.Redirect(failRedirect, http.StatusFound)
		return
	}

	if tokenResp.ErrCode != 0 {
		logs.Error("WeChat access token API error: [%d] %s", tokenResp.ErrCode, tokenResp.ErrMsg)
		c.Redirect(failRedirect, http.StatusFound)
		return
	}

	if tokenResp.AccessToken == "" || tokenResp.OpenID == "" {
		logs.Error("WeChat access token response missing access_token or openid. Body: %s", string(body))
		c.Redirect(failRedirect, http.StatusFound)
		return
	}

	// 4. 使用 access_token 和 openid 获取用户信息
	userInfoURL := fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN",
		tokenResp.AccessToken,
		tokenResp.OpenID,
	)

	userInfoResp, err := http.Get(userInfoURL)
	if err != nil {
		logs.Error("Failed to request WeChat user info: %v", err)
		c.Redirect(failRedirect, http.StatusFound)
		return
	}
	defer userInfoResp.Body.Close()

	userInfoBody, err := io.ReadAll(userInfoResp.Body)
	if err != nil {
		logs.Error("Failed to read WeChat user info response body: %v", err)
		c.Redirect(failRedirect, http.StatusFound)
		return
	}

	type WechatUserInfoResponse struct {
		OpenID     string   `json:"openid"`
		Nickname   string   `json:"nickname"`
		Sex        int      `json:"sex"` // 1为男性，2为女性
		Province   string   `json:"province"`
		City       string   `json:"city"`
		Country    string   `json:"country"`
		HeadImgURL string   `json:"headimgurl"`
		Privilege  []string `json:"privilege"`
		UnionID    string   `json:"unionid"` // 可能为空
		ErrCode    int      `json:"errcode"` // 获取失败时返回
		ErrMsg     string   `json:"errmsg"`  // 获取失败时返回
	}

	var userInfo WechatUserInfoResponse
	if err := json.Unmarshal(userInfoBody, &userInfo); err != nil {
		logs.Error("Failed to unmarshal WeChat user info response: %v. Body: %s", err, string(userInfoBody))
		c.Redirect(failRedirect, http.StatusFound)
		return
	}

	if userInfo.ErrCode != 0 {
		logs.Error("WeChat user info API error: [%d] %s", userInfo.ErrCode, userInfo.ErrMsg)
		c.Redirect(failRedirect, http.StatusFound)
		return
	}

	user := models.User{
		WechatOpenId:  sql.NullString{String: userInfo.OpenID, Valid: true},
		WechatUnionId: sql.NullString{String: userInfo.UnionID, Valid: true},
		Avatar:        userInfo.HeadImgURL,
		Nickname:      userInfo.Nickname,
	}

	if err := user.Get(); err == nil {
		// user exists, nothing to do
	} else {
		// new user, register
		id, err := user.Insert()
		if err != nil {
			logs.Error("register user %s error: %v", user.WechatOpenId, err)
			c.Redirect(failRedirect, http.StatusFound)
		}
		user.Id = int(id)
	}

	c.SetSession(conf.USER_ID, user.Id)

	// 7. 重定向到登录成功页面
	logs.Info("User logged in successfully via WeChat.")
	c.Redirect(successRedirect, http.StatusFound)
}

func (c *UserApiController) UserProfile() {
	userId := c.GetUserId() // 从中间件获取用户ID
	user := models.User{Id: userId}
	if err := user.Get(); err != nil {
		c.ErrorJson(http.StatusInternalServerError, err.Error())
		return
	}
	data := vo.UserInfoVo{
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
	}
	c.JSON(common.JsonResult{
		Code: 0,
		Data: data,
	})
}

func (c *UserApiController) Logout() {
	c.DelSession(conf.USER_ID)
	c.JSON(common.JsonResult{
		Code: 0,
		Msg:  "退出成功",
	})
}

// Login by mobile and password
func (c *UserApiController) MobilePwdLogin() {
	var req dto.UserMobilePwdLoginReq
	if err := c.ParseJSON(&req); err != nil {
		c.ErrorJson(400, "请求参数错误"+err.Error())
		return
	}

	if !sms.IsValidMobile(req.Mobile) {
		c.ErrorJson(400, "无效的手机号码")
		return
	}

	user := models.User{
		Mobile: sql.NullString{String: req.Mobile, Valid: true},
	}
	if err := user.Get(); err != nil {
		// 如果用户不存在，则创建一个新用户
		user.Nickname = DEFAULT_USER_NICKNAME
		user.Avatar = DEFAULT_USER_AVATAR
		id, err := user.Insert()
		if err != nil {
			c.ErrorJson(http.StatusNotFound, err.Error())
			return
		}
		user.Id = int(id)
	}

	if user.Password == "" {
		// c.ErrorJson(400, "用户未设置密码")
		// return
		// TODO: 需要处理用户未设置密码的情况
		user.Password, _ = utils.Md5(DEFAULT_USER_PASSWORD)
	}

	// 密码校验（假设为明文或双MD5，按实际情况调整）
	inputPwd, _ := utils.Md5(req.Password)
	if user.Password != inputPwd {
		c.ErrorJson(401, "密码错误")
		return
	}

	c.SetSession(conf.USER_ID, user.Id)

	c.JSON(common.JsonResult{
		Code: 0,
		Msg:  "登录成功",
		Data: vo.UserInfoVo{
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
		},
	})
}
