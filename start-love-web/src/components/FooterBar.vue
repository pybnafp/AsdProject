<template>
  <div class="foot-container">
    <div class="footer">
      <div>
        <span>{{ copyRight }}</span>
      </div>
    </div>
  </div>
</template>
<script setup>
import { ref } from "vue";

const copyRight = ref("Copyright ©2025  浙ICP备2024126085号-1");
</script>

<style scoped lang="stylus">
.foot-container {
  display flex;
  justify-content center

  .footer {
    font-size: 16px;
  }
}
</style>
