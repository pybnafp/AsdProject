<template>
  <div class="welcome app-background">
    <div class="container">
      <img src="@/assets/img/logo.svg" alt=""/>
      <div class="title">您好，欢迎使用SCAI</div>
      <ChatInput @send="sendMessage"/>
      <div class="active-image">
        <img src="@/assets/img/chat-welcome-bg1.png" alt=""/>
        <img src="@/assets/img/chat-welcome-bg2.png" alt=""/>
      </div>
    </div>
  </div>
</template>
<script setup>
import {onMounted, ref} from "vue";
import ChatInput from "@/components/ChatInput.vue";
import {httpPost} from "@/utils/http";
import {showNotify} from "vant";
import {router} from "@/router";

const title = ref(process.env.VUE_APP_TITLE);

// 新建会话
const createChatMessage = (chatItem) => {
  httpPost("/api/chat/create_stream", chatItem)
      .then((res) => {
        if (res.code === 0) {
          let chatId = res.data.chat_id
          router.push(`/mobile/chat/${chatId}`);
        }
      })
      .catch((e) => {
        showNotify({ message: '创建对话失败！' });
      });
};
// 发送消息
const sendMessage = (chatItem) => {
  createChatMessage(chatItem)
};
onMounted(() => {

});

</script>
<style scoped lang="stylus">
@import "@/assets/css/mobile/add-chat.styl"
</style>
