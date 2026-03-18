<template>
  <div class="single-page">
    <el-container>
      <el-header>
        <header-bar></header-bar>
      </el-header>
      <el-main>
        <div id="container" :style="{ height: mainWinHeight + 'px' }">
          <div class="text-box">
            <img style="width: 128px;height: 128px; margin: auto" src="/images/illustration/bg11.png" alt="">
            <p style="text-align:center; font-size: 18px; line-height: 40px; margin: 10px 0 40px 0">福建省与你孤独症儿童服务中心成立于2025年1月，是福建省残联指导下首家专注孤独症儿童关爱服务的公益慈善组织。中心由一群怀揣公益理想的跨界青年志愿者共同发起，以 “看见、链接、发展”为核心理念，致力于为"星星的孩子"搭建起一座座温暖与希望的桥梁</p>
            <div class="text-img">
              <img style="width: 53%" src="/images/illustration/bg12.png" alt="">
              <img style="width: 45%" src="/images/illustration/bg13.png" alt="">
              <img style="width: 45%" src="/images/illustration/bg14.png" alt="">
              <img style="width: 53%" src="/images/illustration/bg15.png" alt="">
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

.text-box {
  width 80%;
  margin 50px auto;
  .text-img {
    display flex;
    justify-content space-between;
    gap 10px;
    flex-wrap wrap;
  }
}
</style>