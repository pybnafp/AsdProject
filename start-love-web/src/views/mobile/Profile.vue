<template>
  <div class="app-background">
    <div class="container mobile-user-profile">
      <van-nav-bar
          :title="title"
          @click-left="showPicker = true"
          custom-class="navbar">
      </van-nav-bar>
      <div class="content">
        <div class="user-box" v-if="isLogin">
          <div class="user-info">
            <img :src="userInfo.avatar" alt=""/>
            <span class="nickname" style="color:#1B2559;">{{ userInfo.nickname }}</span>
          </div>
          <i class="iconfont icon-kaiguan icon-right" @click="logout"></i>
        </div>
        <div class="user-box" v-else>
          <div class="user-info" @click="goLogin">
            <img src="@/assets/img/avatar.png" alt=""/>
            <span class="nickname">登录/注册</span>
          </div>
          <img class="icon-right" src="@/assets/img/icon-youjiantou.png" alt=""/>
        </div>
        <div class="menu-list-first">
          <a :href="item.href" v-for="item in menusFirst" :key="item.href">
            <span>{{ item.title }}</span>
            <img src="@/assets/img/icon-arrows-right.png" alt=""/>
          </a>
        </div>
        <div class="menu-list-second">
          <a :href="item.href" v-for="item in menusSecond" :key="item.href">
            <span>{{ item.title }}</span>
            <img src="@/assets/img/icon-arrows-right.png" alt=""/>
          </a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import {onMounted, ref} from "vue";
import {httpPost} from "@/utils/http";
import {ElMessage} from "element-plus";
import {checkSession} from "@/store/cache";
import {useRouter} from "vue-router";
import {useSharedStore} from "@/store/sharedata";

const title = ref("我的");
const router = useRouter();
const isLogin = ref(false);
const store = useSharedStore();
const userInfo = store.getUserInfo;
const showPicker = ref(false)
const menusFirst = ref([
  {
    title: "早筛技术",
    href: "/mobile/single-page/early-screening-technology",
  },
  {
    title: "政策指引",
    href: "/mobile/single-page/policy-guide",
  },
  {
    title: "星启协爱",
    href: "/mobile/single-page/star-love",
  },
  {
    title: "公益机构",
    href: "/mobile/single-page/public-institution",
  }
]);
const menusSecond = ref([
  {
    title: "开放平台",
    href: "/mobile/single-page/open-platform",
  },
  {
    title: "关于我们",
    href: "/mobile/single-page/about",
  }
]);

onMounted(() => {
  checkSession()
      .then((res) => {
        isLogin.value = true;
      })
      .catch(() => {
      });

});
const goLogin = () => {
  router.push("/mobile/login");
}
const logout = function () {
  httpPost("/api/logout")
      .then(() => {
        store.setUserInfo({})
        store.setIsLogin(false);
        router.push("/mobile/login");
      })
      .catch(() => {
        ElMessage.error("注销失败！");
        router.push("/mobile/login");
      });
};
</script>

<style lang="stylus" scoped>
@import "@/assets/css/mobile/profile.styl"
</style>
