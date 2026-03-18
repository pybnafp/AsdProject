<template>
  <div class="welcome">
    <div class="container">
      <div class="title">您好，欢迎使用SCAI</div>
      <ChatInput @send="send"/>
      <div class="active-image">
        <img src="@/assets/img/chat-welcome-bg1.png" alt="" />
        <img src="@/assets/img/chat-welcome-bg2.png" alt="" />
      </div>
    </div>
  </div>
</template>
<script setup>
import { onMounted, ref } from "vue";
import ChatInput from "@/components/ChatInput.vue";

const title = ref(process.env.VUE_APP_TITLE);

const emits = defineEmits(["send"]);
const send = (chatItem) => {
  emits("send", chatItem);
};
onMounted(() => {

});

</script>
<style scoped lang="stylus">
.welcome {
  text-align center
  display flex
  justify-content center
  margin-top 10vh

  .container {
    max-width 768px;
    width: 100%;

    .title {
      font-size: 32px;
      line-height: 100%;
      text-align: center;
      margin-bottom: 44px;
    }
    .active-image {
      display: flex;
      justify-content: space-between;
      margin-top: 30px;
      img{
        width: 365px;
        height: auto;
        border-radius: 40px;
        background: rgba(93, 25, 210, 0.05);
        backdrop-filter: blur(9.3px);
        padding :8px;
      }
    }
  }
}
</style>
