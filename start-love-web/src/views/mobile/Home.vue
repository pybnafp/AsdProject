<template>
  <van-config-provider :theme="theme">
    <div class="mobile-home">
      <router-view/>
      <van-tabbar route v-model="active">
        <van-tabbar-item to="/mobile/chat/add" name="chat">
          <span>新建会话</span>
          <template #icon="props">
            <img :src="props.active ? icon.add.active :  icon.add.normal"  alt=""/>
          </template>
        </van-tabbar-item>
        <van-tabbar-item to="/mobile/chat/list" name="historyChat">
          <span>历史会话</span>
          <template #icon="props">
            <img :src="props.active ? icon.chat.active :  icon.chat.normal"  alt=""/>
          </template>
        </van-tabbar-item>
        <van-tabbar-item to="/mobile/report" name="report">
          <span>我的报告</span>
          <template #icon="props">
            <img :src="props.active ? icon.file.active :  icon.file.normal"  alt=""/>
          </template>
        </van-tabbar-item>
        <van-tabbar-item to="/mobile/profile" name="more">
          <span>更多</span>
          <template #icon="props">
            <img :src="props.active ? icon.more.active :  icon.more.normal"  alt=""/>
          </template>
        </van-tabbar-item>
      </van-tabbar>

    </div>
  </van-config-provider>

</template>

<script setup>
import {ref, watch} from "vue";
import {useSharedStore} from "@/store/sharedata";

const active = ref('index')
const store = useSharedStore()
const theme = ref(store.theme)

const icon = ref({
  add: {
    normal: '/images/menu/icon-add.png',
    active: '/images/menu/icon-add-selected.png',
  },
  chat: {
    normal: '/images/menu/icon-chat.png',
    active: '/images/menu/icon-chat-selected.png',
  },
  file: {
    normal: '/images/menu/icon-file.png',
    active: '/images/menu/icon-file-selected.png',
  },
  more: {
    normal: '/images/menu/icon-more.png',
    active: '/images/menu/icon-more-selected.png',
  },
})

watch(() => store.theme, (val) => {
  theme.value = val
})

</script>

<style lang="stylus">
//@import '@/assets/iconfont/iconfont.css';
.mobile-home {
  .container {
    .van-nav-bar {
      position fixed
      width 100%
    }

    padding 0 10px
  }

}

// 黑色主题
.van-theme-dark body {
  background #1c1c1e
}

.van-nav-bar {
  position fixed
  width 100%
}
</style>