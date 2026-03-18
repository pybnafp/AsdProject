<template>
  <el-config-provider>
    <router-view />
  </el-config-provider>
</template>

<script setup>
import { ElConfigProvider } from "element-plus";
import { onMounted, watch } from "vue";
import { isChrome, isMobileV2 } from "@/utils/libs";
import { showMessageInfo } from "@/utils/dialog";
import { useSharedStore } from "@/store/sharedata";

const debounce = (fn, delay) => {
  let timer;
  return (...args) => {
    if (timer) {
      clearTimeout(timer);
    }
    timer = setTimeout(() => {
      fn(...args);
    }, delay);
  };
};

const _ResizeObserver = window.ResizeObserver;
window.ResizeObserver = class ResizeObserver extends _ResizeObserver {
  constructor(callback) {
    callback = debounce(callback, 200);
    super(callback);
  }
};

const store = useSharedStore();
onMounted(() => {
  if (!isChrome() && !isMobileV2()) {
    showMessageInfo("建议使用 Chrome 浏览器以获得最佳体验。");
  }

  // 设置主题
  document.documentElement.setAttribute("data-theme", store.theme);
});

watch(
  () => store.isLogin,
  (val) => {
    if (val) {
    }
  }
);
</script>

<style lang="stylus">
html, body {
  margin: 0;
  padding: 0;
}

#app {
  margin: 0 !important;
  padding: 0 !important;
  font-family: Helvetica Neue, Helvetica, PingFang SC, Hiragino Sans GB, Microsoft YaHei, Arial, sans-serif
  -webkit-font-smoothing: antialiased;
  text-rendering: optimizeLegibility;

  --primary-color: #21aa93

  h1 { font-size: 2em; } /* 通常是 2em */
  h2 { font-size: 1.5em; } /* 通常是 1.5em */
  h3 { font-size: 1.17em; } /* 通常是 1.17em */
  h4 { font-size: 1em; } /* 通常是 1em */
  h5 { font-size: 0.83em; } /* 通常是 0.83em */
  h6 { font-size: 0.67em; } /* 通常是 0.67em */

}

.el-overlay-dialog {
  display flex
  justify-content center
  align-items center
  overflow hidden

  .el-dialog {
    margin 0;

    .el-dialog__body {
      //max-height 80vh
      overflow-y auto
    }
  }
}

/* 省略显示 */
.ellipsis {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.van-toast--fail {
  background #fef0f0
  color #f56c6c
}

.van-toast--success {
  background #D6FBCC
  color #07C160
}

//@import '@/assets/iconfont/iconfont.css'
</style>
