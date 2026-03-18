<template>
  <div class="single-page">
    <el-container>
      <el-header>
        <header-bar></header-bar>
      </el-header>
      <el-main>
        <div id="container" :style="{ height: mainWinHeight + 'px' }">
          <div class="circle-box">
            <img src="/images/illustration/bg10.png" alt="星球人的圈子">
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

.circle-box {
  width 80%;
  margin 50px auto;
}
</style>