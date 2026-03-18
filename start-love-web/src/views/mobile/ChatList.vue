<template>
  <div class="app-background">
    <div class="container mobile-chat-list">
      <van-nav-bar
          :title="title"
          @click-left="showPicker = true"
          custom-class="navbar"
      >
      </van-nav-bar>
      <div class="content">
        <van-list
            v-model:error="error"
            v-model:loading="loading"
            :finished="finished"
            error-text="请求失败，点击重新加载"
            finished-text="没有更多了"
            @load="onLoad"
        >
          <van-swipe-cell v-for="item in chats" :key="item.id">
            <van-cell @click="changeChat(item)">
              <div class="chat-list-item">
                <div class="chat-list-text">
                  <img src="@/assets/img/icon-chat-avatar.svg" alt=""/>
                  <div class="van-ellipsis">{{ item.title }}</div>
                </div>
                <img src="@/assets/img/icon-arrows.png" alt=""/>
              </div>
            </van-cell>
            <template #right>
              <van-button square text="修改" type="primary" @click="editChat(item)"/>
              <van-button square text="删除" type="danger" @click="removeChat(item)"/>
            </template>
          </van-swipe-cell>
        </van-list>
      </div>
    </div>
  </div>
</template>

<script setup>
import {ref} from "vue";
import { httpPost} from "@/utils/http";
import {showFailToast} from "vant";
import {checkSession} from "@/store/cache";
import {router} from "@/router";

const title = ref("会话列表")
const chats = ref([])
const loading = ref(false)
const finished = ref(false)
const error = ref(false)
const loginUser = ref(null)
const isLogin = ref(false)
const showPicker = ref(false)
const offset = ref(0)
const limit = ref(10)

checkSession().then((user) => {
  loginUser.value = user
  isLogin.value = true

}).catch(() => {
  loading.value = false
  finished.value = true
})

const onLoad = () => {
  httpPost("/api/chat/list", {
    offset: offset.value,
    limit: limit.value,
  }).then((res) => {
    loading.value = false;
    if (res.data) {
      const items = res.data;

      if (offset.value === 0) {
        chats.value = items;
      } else {
        chats.value = [...chats.value, ...items];
      }

      if (chats.value.length < res.count) {
        offset.value = offset.value + limit.value;
      } else {
        finished.value = true
      }
    } else {
      finished.value = true
    }
  }).catch(() => {
    loading.value = false;
    error.value = true
    showFailToast("加载会话列表失败")
  })
};

const changeChat = (chat) => {
  router.push(`/mobile/chat/${chat.chat_id}`)
}

const editChat = (row) => {
}

const removeChat = (item) => {

}

</script>

<style lang="stylus" scoped>
@import "@/assets/css/mobile/chat-list.styl"
</style>