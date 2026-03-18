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
      <div class="circle-box">
        <img src="/images/illustration/bg18.png" alt="星球人的圈子">
      </div>
    </div>
  </div>
</template>
<script setup>
import {onMounted, ref} from 'vue';

const title = ref("星启协爱");
const showPicker = ref(false)

onMounted(() => {
});
const goBack = () => {
  history.back();
}

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
.circle-box {
  width 100%;
  margin 50px auto;
}
</style>