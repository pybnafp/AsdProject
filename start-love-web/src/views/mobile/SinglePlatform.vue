<template>
  <div class="app-background">
    <van-nav-bar
        :title="title"
        @click-left="showPicker = true"
        custom-class="navbar"
    >
      <template #left>
        <img style="width: 28px;height: 28px" src="@/assets/img/icon-arrows-left.png" alt="" @click="goBack"/>
      </template>
    </van-nav-bar>
    <div id="container">
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
</template>
<script setup>
import {onMounted, ref} from 'vue';

const title = ref("开放平台");
const showPicker = ref(false)
const goBack = () => {
  history.back();
}

onMounted(() => {
});

</script>

<style lang="stylus">
.app-background {
  height: 100%;

  #container {
    min-width: 0;
    flex: 1;
    padding: 46px 15px;
  }
}

.content-item {
  margin-bottom 24px;

  .title {
    font-size: 12px;
    line-height: 20px;
    color: #999797;
    margin-bottom 5px;
  }

  .logo-list {
    border-radius: 12px;
    background: #FFFFFF;
    box-shadow: 0 4px 40px 0 #F0EDF6;
    padding 10px;
    display flex;
    justify-content space-evenly;
    flex-wrap wrap;


    img {
      height 60px;
    }
  }

}
</style>