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
      <div class="text-box">
        <img style="width: 72px;height: 72px; margin: auto" src="/images/illustration/bg11.png" alt="">
        <p style="font-size: 14px; line-height: 24px; margin: 15px 0; text-align: center">
          福建省与你孤独症儿童服务中心成立于2025年1月，是福建省残联指导下首家专注孤独症儿童关爱服务的公益慈善组织。中心由一群怀揣公益理想的跨界青年志愿者共同发起，以
          “看见、链接、发展”为核心理念，致力于为"星星的孩子"搭建起一座座温暖与希望的桥梁</p>
        <div class="text-img">
          <img src="/images/illustration/bg12.png" alt="">
          <img src="/images/illustration/bg13.png" alt="">
          <img src="/images/illustration/bg14.png" alt="">
          <img src="/images/illustration/bg15.png" alt="">
        </div>
      </div>
    </div>
  </div>
</template>
<script setup>
import {onMounted, ref} from 'vue';

const title = ref("公益机构");
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
.text-box {
  margin 30px auto;

  .text-img {
    width 100%;
    img {
      width 100%;
      margin-bottom 10px;
    }
  }
}
</style>