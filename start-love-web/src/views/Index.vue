<template>
  <div class="index-page">
    <div class="menu-box">
      <el-menu mode="horizontal" :ellipsis="false">
        <div class="menu-item">
          <img :src="logo" class="logo" alt="星启诶艾" />
        </div>
        <div class="menu-item">
          <span v-if="!isLogin">
            <el-button @click="router.push('/login')" class="btn-go animate__animated animate__pulse animate__infinite" round>登录/注册</el-button>
          </span>
        </div>
      </el-menu>
    </div>
    <div class="content">
      <div style="height: 158px"></div>
      <h1 class="animate__animated animate__backInDown">
        {{ title }}
      </h1>
      <h2>
        {{ slogan }}
      </h2>
      <div class="navs animate__animated animate__backInDown">
        <el-space wrap :size="14">
          <div v-for="item in navs" :key="item.url" class="nav-item-box" @click="router.push(item.url)">
            <i :class="'iconfont ' + iconMap[item.url]"></i>
            <div>{{ item.name }}</div>
          </div>
        </el-space>
      </div>
    </div>

    <footer-bar />
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import FooterBar from "@/components/FooterBar.vue";
import { checkSession } from "@/store/cache";
import { isMobileV2 } from "@/utils/libs";

const router = useRouter();

if (isMobileV2()) {
  router.push("/mobile/index");
}

const title = ref(process.env.VUE_APP_TITLE);
const logo = ref("/images/logo.svg");
const slogan = ref("");
const isLogin = ref(false);

onMounted(() => {
  checkSession()
    .then(() => {
      isLogin.value = true;
    })
    .catch(() => {});
});
</script>

<style lang="stylus" scoped>
@import "@/assets/css/index.styl"
</style>
