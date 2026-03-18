<template>
  <el-container class="send-verify-code">
    <el-button type="success" :size="props.size" :disabled="!canSend" @click="sendMsg" style="width: 100px">
      {{ btnText }}
    </el-button>
  </el-container>
</template>

<script setup>
// 发送短信验证码组件
import { ref } from "vue";
import { validateEmail, validateMobile } from "@/utils/validate";
import { ElMessage } from "element-plus";
import {httpPost} from "@/utils/http";

// eslint-disable-next-line no-undef
const props = defineProps({
  receiver: String,
  size: String,
  type: {
    type: String,
    default: "mobile",
  },
});
const btnText = ref("发送验证码");
const canSend = ref(true);

const sendMsg = () => {
  if (!validateMobile(props.receiver) && props.type === "mobile") {
    return ElMessage.error("请输入合法的手机号");
  }
  if (!validateEmail(props.receiver) && props.type === "email") {
    return ElMessage.error("请输入合法的邮箱地址");
  }
  doSendMsg({})
};

const doSendMsg = (data) => {
  if (!canSend.value) {
    return;
  }
  //测试代码
  /*canSend.value = false;
  ElMessage.success("验证码发送成功");
  let time = 60;
  btnText.value = time;
  const handler = setInterval(() => {
    time = time - 1;
    if (time <= 0) {
      clearInterval(handler);
      btnText.value = "重新发送";
      canSend.value = true;
    } else {
      btnText.value = time;
    }
  }, 1000);*/

  httpPost("/api/login/send-code", {
    mobile: props.receiver,
  })
    .then(() => {
      ElMessage.success("验证码发送成功");
      let time = 60;
      btnText.value = time;
      const handler = setInterval(() => {
        time = time - 1;
        if (time <= 0) {
          clearInterval(handler);
          btnText.value = "重新发送";
          canSend.value = true;
        } else {
          btnText.value = time;
        }
      }, 1000);
    })
    .catch((e) => {
      canSend.value = true;
      ElMessage.error("验证码发送失败：" + e.message);
    });
};
</script>

<style lang="stylus" scoped>

.send-verify-code {
  .send-btn {
    width: 100%;
  }
}
</style>
