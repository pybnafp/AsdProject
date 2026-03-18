<template>
  <div class="flex-center login-page" :style="{'height':mainHeight + 'px'}">
    <div class="login-box">
      <el-row class="login-container">
        <el-col :span="12" class="welcome-bg">
          <el-image src="/images/login-welcome-bg.png" fit="cover"/>
        </el-col>
        <el-col :span="12" class="login-box">
          <div class="corner-mark">
            <img src="@/assets/img/icon-qrcode.svg" alt="" v-if="loginWay === 'account'" @click="changeLoginWay"/>
            <img src="@/assets/img/icon-pc.svg" alt="" v-else @click="changeLoginWay"/>
          </div>
          <div class="login-tip">
            <p class="title">登录后免费使用完整功能</p>
            <p class="sub-title">您好，欢迎使用SCAI</p>
          </div>
          <div class="input-form" v-if="loginWay === 'account'">
            <el-form ref="ruleFormRef" :model="ruleForm" :rules="rules">
              <el-form-item label="" prop="mobile">
                <el-input v-model="ruleForm.mobile" size="large" placeholder="请输入手机号" @keyup="handleKeyup">
                  <template #prepend>
                    <el-select size="large" v-model="ruleForm.areaCode" placeholder="+86" style="width: 100px">
                      <el-option label="+86" value="86"/>
                      <el-option label="+21" value="21"/>
                    </el-select>
                  </template>
                </el-input>
              </el-form-item>
              <el-form-item label="" prop="password" v-if="loginType === 'password'">
                <el-input size="large" v-model="ruleForm.password" placeholder="请输入密码" type="password"
                          show-password
                          autocomplete="off" @keyup="handleKeyup"/>
              </el-form-item>
              <el-form-item label="" prop="captcha" v-else>
                <el-input size="large" v-model="ruleForm.captcha" placeholder="请输入验证码" autocomplete="off"
                          @keyup="handleKeyup">
                  <template #append>
                    <SendMsg size="large" :receiver="ruleForm.mobile" type="mobile"/>
                  </template>
                </el-input>
              </el-form-item>
              <el-checkbox v-model="ruleForm.isAgree" size="large">
                <el-text type="info">
                  已阅读并同意 {{ title }} 的
                  <el-link href="https://baidu.com" target="_blank" type="primary" style="vertical-align: baseline">
                    使用协议
                  </el-link>
                  和
                  <el-link href="https://baidu.com" target="_blank" type="primary" style="vertical-align: baseline">
                    隐私政策
                  </el-link>
                </el-text>
              </el-checkbox>
              <el-form-item style="margin-top: 60px">
                <el-button class="login-btn" size="large" type="primary" @click="login">立即登录</el-button>
              </el-form-item>
              <el-form-item v-if="loginType === 'captcha'">
                <el-button class="login-btn" plain size="large" type="primary" @click="changeLoginType">密码登录
                </el-button>
              </el-form-item>
              <el-form-item v-else>
                <el-button class="login-btn" plain size="large" type="primary" @click="changeLoginType">验证码登录
                </el-button>
              </el-form-item>

            </el-form>
          </div>
          <div class="qrcode-box" v-else>
            <div class="qrcode-img" v-loading="qrcodeLoading">
              <wxlogin
                  v-if="wechatConfig.appid && wechatConfig.redirect_uri"
                  :appid="wechatConfig.appid"
                  scope="snsapi_login"
                  :redirect_uri="wechatConfig.redirect_uri"
                  :href="wechatConfig.href"
                  :state="wechatConfig.state"
              ></wxlogin>
            </div>
            <div class="tip">
              <el-text type="info">打开微信APP - 点击右上角加号 - 点击扫一扫</el-text>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup>
import {onMounted, ref, reactive} from "vue";
import {httpPost} from "@/utils/http";
import {useRouter} from "vue-router";
import {isMobileV2} from "@/utils/libs";
import {showMessageError} from "@/utils/dialog";
import {useSharedStore} from "@/store/sharedata";
import wxlogin from 'vue-next-wxlogin';
import SendMsg from "@/components/SendMsg.vue";

const router = useRouter();
const title = ref(process.env.VUE_APP_TITLE);
const store = useSharedStore();

const loginType = ref("captcha");
const loginWay = ref("account");
const qrcodeLoading = ref(true);
const enableVerify = ref(false);
const captchaRef = ref(null);
const ruleFormRef = ref(null);
const ruleForm = reactive({
  areaCode: "86",
  mobile: process.env.VUE_APP_USER,
  password: process.env.VUE_APP_PASS,
  captcha: "",
  isAgree: true,
});
const rules = {
  mobile: [{required: true, trigger: "blur", message: "请输入手机号"}],
  password: [{required: true, trigger: "blur", message: "请输入密码"}],
  captcha: [{required: true, trigger: "blur", message: "请输入验证码"}],
  isAgree: [{required: true, trigger: "blur", message: "请阅读并同意"}],
};
const wechatConfig = ref({
  appid: "",
  redirect_uri: "",
  state: "",
  href: process.env.VUE_APP_API_HOST + "/css/wxlogin.css",
})
const mainHeight = ref(0);
const resizeElement = function () {
  mainHeight.value = window.innerHeight;
};
onMounted(() => {
  resizeElement()
  if (store.isLogin) {
    if (isMobileV2()) {
      router.push("/mobile");
    } else {
      router.push("/chat");
    }
  }
});
const changeLoginType = () => {
  if (loginType.value === "captcha") {
    loginType.value = "password";
  } else {
    loginType.value = "captcha";
  }
}
const changeLoginWay = () => {
  if (loginWay.value === "account") {
    loginWay.value = "qrcode";
    getWechatConfig();
  } else {
    loginWay.value = "account";
  }
}
const handleKeyup = (e) => {
  if (e.key === "Enter") {
    login();
  }
};
const getWechatConfig = () => {
  /*setTimeout(function () {
    wechatLoginUrl.value = '/images/qrcode.png';
    qrcodeLoading.value = false;
  }, 3000)*/
  httpPost("/api/login/wechat")
      .then((res) => {
        if (res.code === 0) {
          wechatConfig.value.appid = res.data.appid;
          wechatConfig.value.state = res.data.state;
          wechatConfig.value.redirect_uri = res.data.redirect_uri;
          qrcodeLoading.value = false;
        } else {
          showMessageError("登录失败，" + res.msg);
        }
      })
      .catch((e) => {
        showMessageError("登录失败，" + e.message);
      });
};
const login = async () => {
  await ruleFormRef.value.validate(async (valid) => {
    if (valid) {
      if (loginType.value === "captcha") {
        doCaptchaLogin()
      } else {
        doPasswordLogin()
      }
    }
  });
};
const doPasswordLogin = () => {
  httpPost("/api/login/mobile-pwd", {
    mobile: ruleForm.mobile,
    password: ruleForm.password,
  })
      .then((res) => {
        if (res.code !== 0) {
          showMessageError(res.msg);
          return;
        }
        store.setUserInfo(res.data);
        store.setIsLogin(true);
        if (isMobileV2()) {
          router.push("/mobile");
        } else {
          router.push("/chat");
        }
      })
      .catch((e) => {
        showMessageError("登录失败，" + e.message);
      });
};
const doCaptchaLogin = () => {
  httpPost("/api/login/mobile", {
    mobile: ruleForm.mobile,
    passcode: ruleForm.captcha,
  })
      .then((res) => {
        if (res.code !== 0) {
          showMessageError(res.msg);
          return;
        }
        store.setUserInfo(res.data);
        store.setIsLogin(true);
        if (isMobileV2()) {
          router.push("/mobile");
        } else {
          router.push("/chat");
        }
      })
      .catch((e) => {
        showMessageError("登录失败，" + e.message);
      });
};
</script>

<style lang="stylus" scoped>
@import "@/assets/css/login.styl"
</style>
