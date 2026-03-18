<template>
  <div class="chat-page">
    <el-container>
      <el-header>
        <div class="menu-container">
          <header-bar></header-bar>
        </div>
      </el-header>
      <el-main>
        <div class="chat-container">
          <div id="container" :style="{ height: mainWinHeight + 'px' }">
            <div class="chat-box" id="chat-box" :style="{ height: chatBoxHeight + 'px' }">
              <div v-if="showHello">
                <Welcome @send="sendMessage"/>
              </div>
              <div class="chat-messages" v-for="item in messages" :key="item.message_id" v-else>
                <chat-prompt :data="item"/>
                <chat-reply :data="item"
                            :thinking="thinking && (item.message_id === currentMessage.message_id)"
                            :loading="loading && (item.message_id === currentMessage.message_id)"
                />
              </div>
            </div>
            <div class="input-container" v-if="!showHello">
              <ChatInput @send="sendMessage"/>
            </div>
          </div>
        </div>
      </el-main>
      <el-footer v-if="showHello">
        <FooterBar/>
      </el-footer>
    </el-container>
  </div>
</template>
<script setup>
import {inject, nextTick, onMounted, onUnmounted, ref, watch} from "vue";
import ChatPrompt from "@/components/ChatPrompt.vue";
import ChatReply from "@/components/ChatReply.vue";
import "highlight.js/styles/a11y-dark.css";
import {isMobileV2} from "@/utils/libs";
import {ElMessage} from "element-plus";
import {httpPost} from "@/utils/http";
import {useRouter} from "vue-router";
import Clipboard from "clipboard";
import Welcome from "@/components/Welcome.vue";
import FooterBar from "@/components/FooterBar.vue";
import ChatInput from "@/components/ChatInput.vue";
import {fetchEventSource} from "@microsoft/fetch-event-source";
import HeaderBar from "@/components/HeaderBar.vue";

const chatDetail = ref({});
const mainWinHeight = ref(0); // 主窗口高度
const chatBoxHeight = ref(0); // 聊天内容框高度
const router = useRouter();
const chatId = ref("");
const showHello = ref(true);
const refreshChatList = inject("refreshChatList");
const messages = ref([]);
const thinking = ref(false);
const lineBufferReasoning = ref(""); // 思考输出缓冲行
const lineBufferAnswer = ref(""); // 答案输出缓冲行
const loading = ref(false);
const currentMessage = ref({});

if (isMobileV2()) {
  if (chatId.value) {
    router.push("/mobile/chat/" + chatId.value);
  } else {
    router.push("/mobile/chat/add");
  }
}

onMounted(() => {
  initData();
  resizeElement();
  window.onresize = () => resizeElement()

  const clipboard = new Clipboard(".copy-reply, .copy-code-btn");
  clipboard.on("success", () => {
    ElMessage.success("复制成功！");
  });

  clipboard.on("error", () => {
    ElMessage.error("复制失败！");
  });

  scrollChatBox();
});

onUnmounted(() => {

});
// 初始化数据
const initData = () => {
  // 初始化 ChatID
  chatId.value = router.currentRoute.value.params.id || "";
  if (chatId.value) {
    showHello.value = false;
    // 查询对话信息
    httpPost("/api/chat/detail", {chat_id: chatId.value})
        .then((res) => {
          chatDetail.value = res.data
          messages.value = res.data.messages
          if (messages.value[messages.value.length - 1]['completion'] === '') {
            currentMessage.value = messages.value[messages.value.length - 1];
            readChatMessage(messages.value[messages.value.length - 1]['message_id'])
          }
        })
        .catch((e) => {
          console.error("获取对话信息失败：" + e.message);
        });
  } else {
    console.log(showHello.value);
  }
};
const scrollChatBox = () => {
  setTimeout(() => {
    const chatBox = document.getElementById("chat-box");
    if (chatBox) {
      chatBox.scrollTo(0, chatBox.scrollHeight);
    }
  }, 500);
}
const resizeElement = function () {
  if (chatId.value) {
    mainWinHeight.value = window.innerHeight - 84;
    chatBoxHeight.value = window.innerHeight - 160;
  } else {
    mainWinHeight.value = window.innerHeight - 114;
    chatBoxHeight.value = window.innerHeight - 100;
  }
};

// 新建会话
const createChatMessage = (chatItem) => {
  httpPost("/api/chat/create_stream", chatItem)
      .then((res) => {
        if (res.code === 0) {
          if (chatId.value) {
            // 追加消息
            messages.value.push(res.data.message);
            currentMessage.value = res.data.message;
            scrollChatBox();
            readChatMessage(res.data.message.message_id)
          } else {
            refreshChatList();
            let chatId = res.data.chat_id
            router.push(`/chat/${chatId}`);
          }
        }
      })
      .catch((e) => {
        ElMessage.error("创建对话失败！");
      });
};
// 读取会话
const readChatMessage = async (messageId) => {
  thinking.value = true;
  loading.value = true;
  const ctrl = new AbortController(); // 用于中断请求
  await fetchEventSource('/api/chat/read_stream', {
    credentials: 'include',
    headers: {
      "Content-Type": "application/json",
      Accept: ["text/event-stream"],
    },
    body: JSON.stringify({
      chat_id: chatId.value,
      message_id: messageId,
    }),
    method: 'POST',
    signal: ctrl.signal,
    openWhenHidden: true, // 页面退至后台时保持连接
    onopen: (response) => {
      console.log('打开连接', response)
    },
    onmessage: (event) => {
      if (event.data) {
        let data = JSON.parse(event.data);
        const reply = messages.value[messages.value.length - 1];
        if (data.type === "reasoning") {
          loading.value = false;
          lineBufferReasoning.value += data.content;
          reply["reasoning"] = lineBufferReasoning.value;
        }else if (data.type === "answer") {
          loading.value = false;
          thinking.value = false;
          lineBufferAnswer.value += data.content;
          reply["completion"] = lineBufferAnswer.value;
        } else if (data.type === "end") {
          lineBufferReasoning.value = ""; // 清空缓冲
          lineBufferAnswer.value = "";
          reply["completion"] = data.content;
        }
        scrollChatBox();
      }
    },
    onerror: (error) => {
      loading.value = false;
      ctrl.abort(); // 中断请求
    },
  })
};

// 发送消息
const sendMessage = (chatItem) => {
  chatItem.chat_id = chatId.value;
  // todo push 消息到会话列表
  createChatMessage(chatItem)
};

</script>

<style scoped lang="stylus">
@import "@/assets/css/chat.styl"
</style>

<style lang="stylus">
.notice-dialog {
  .el-dialog__header {
    padding-bottom 0
  }

  .el-dialog__body {
    padding 0 20px

    ol, ul {
      padding-left 10px
    }

    ol {
      list-style decimal-leading-zero
      padding-left 20px
    }

    ul {
      list-style disc
    }
  }
}

.input-container {
  .el-textarea {
    .el-textarea__inner {
      padding-right 40px
    }
  }
}
.el-loading-spinner {
  text-align left;
}
</style>
