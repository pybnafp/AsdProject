<template>
  <div class="single-page">
    <el-container>
      <el-header>
        <header-bar></header-bar>
      </el-header>
      <el-main>
        <div id="container" :style="{ height: mainWinHeight + 'px' }">
          <div class="content-box">
            <div class="content-item">
              <p class="title">指导部门</p>
              <div class="logo-list">
                <img src="/images/institution/canjirenlianhehui.png" alt="">
                <img src="/images/institution/funvlianhehui.png" alt="">
                <img src="/images/institution/jiankangweiyuanhui.png" alt="">
                <img src="/images/institution/kexuejishuting.png" alt="">
              </div>
            </div>
            <div class="content-item">
              <p class="title">高校院所</p>
              <div class="logo-list">
                <img src="/images/institution/yixueyanjiusuo.png" alt="">
                <img src="/images/institution/jiaotongdaxue.png" alt="">
                <img src="/images/institution/zhejiangdaxue.png" alt="">
                <img src="/images/institution/fuzhoudaxue.png" alt="">
                <img src="/images/institution/zhijiang.png" alt="">
                <img src="/images/institution/gewumingde.png" alt="">
                <img src="/images/institution/yixuegaodeng.png" alt="">
              </div>
            </div>
            <div class="content-item">
              <p class="title">医疗单位</p>
              <div class="logo-list">
                <img src="/images/institution/fjmda.png" alt="">
                <img src="/images/institution/ertongyiyuan.png" alt="">
                <img src="/images/institution/fuyoubaojianyuan.png" alt="">
              </div>
            </div>
            <div class="content-item">
              <p class="title">社会机构</p>
              <div class="logo-list">
                <img src="/images/institution/yuniguduzheng.png" alt="">
                <img src="/images/institution/fazhanjijinhui.png" alt="">
                <img src="/images/institution/guanailianmeng.png" alt="">
                <img src="/images/institution/chengshiyanjiuyuan.png" alt="">
              </div>
            </div>
          </div>
        </div>
      </el-main>
    </el-container>
  </div>
</template>
<script setup>
import {onMounted, ref} from 'vue';
import HeaderBar from "@/components/HeaderBar.vue";

const mainWinHeight = ref(0); // 主窗口高度

onMounted(() => {
  resizeElement();
});

const resizeElement = function () {
  mainWinHeight.value = window.innerHeight - 114;
};
</script>

<style lang="stylus">
@import "@/assets/css/single-common.styl"

.content-box {
  width 80%;
  margin 50px auto;
  .content-item {
    margin-bottom 64px;
    .title {
      font-size: 16px;
      line-height: 40px;
      color: #999797;
    }
    .logo-list {
      border-radius: 16px;
      background: #FFFFFF;
      box-shadow: 0 4px 40px 0 #F0EDF6;
      display flex;
      flex-wrap wrap;
      img {
        height 120px;
      }
    }
  }
}
</style>