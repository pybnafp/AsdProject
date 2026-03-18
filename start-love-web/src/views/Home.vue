<template>
  <div class="common-layout">
    <el-container>
      <div class="spread-button" v-show="isCollapse" @click="isCollapse = !isCollapse">
        <img src="@/assets/img/icon-collapse.svg" alt="" />
      </div>
      <el-aside width="300px" v-show="!isCollapse">
        <div class="left-main-box">
          <div class="logo-box">
            <img class="logo" src="@/assets/img/logo.svg" alt="星启诶艾" @click="router.push('/chat')" />
            <img class="icon-collapse" src="@/assets/img/icon-collapse.svg" alt="" @click="isCollapse = !isCollapse" />
          </div>
          <div class="menu">
            <div class="menu-item" @click="myReport" :class="'/report' === curPath ? 'active' : ''">
              <i class="iconfont icon-wenben"></i>
              <span>我的报告</span>
            </div>
            <div class="menu-item" @click="newChat" :class="'/chat' === curPath ? 'active' : ''">
              <i class="iconfont icon-yuanxingtianjia"></i>
              <span>新建会话</span>
            </div>
          </div>
          <div class="chat-history-box">
            <span class="title">历史会话</span>
            <el-scrollbar :height="chatListHeight" v-infinite-scroll="getChatList" :infinite-scroll-disabled="loading">
              <div class="chat-list" >
                <div class="chat-item" :class="chatId === chat.chat_id ? 'active' : ''" v-for="chat in chatList" :key="chat.chat_id" @click="loadChat(chat)">
                  <img src="@/assets/img/icon-chat-avatar.svg" alt="" />
                  <span class="chat-title">{{chat.title}}</span>
                </div>
              </div>
            </el-scrollbar>
          </div>
          <div class="bottom-box" v-if="store.isLogin">
            <div class="user-info">
              <img :src="userInfo.avatar" alt="" />
              <span class="nickname" style="color:#1B2559;">{{userInfo.nickname}}</span>
            </div>
            <i class="iconfont icon-kaiguan icon-right" @click="logout"></i>
          </div>
          <div class="bottom-box" v-else>
            <div class="user-info">
              <img src="@/assets/img/avatar.png" alt="" />
              <span class="nickname">登录/注册</span>
            </div>
            <img class="icon-right" src="@/assets/img/icon-youjiantou.png" alt="" @click="goLogin" />
          </div>
        </div>
      </el-aside>
      <el-main>
        <div class="content">
          <router-view :key="routerViewKey" v-slot="{ Component }">
            <transition name="move" mode="out-in">
              <component :is="Component" :key="curPath"></component>
            </transition>
          </router-view>
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import { useRouter } from "vue-router";
import { onMounted, provide, ref,} from "vue";
import { httpPost } from "@/utils/http";
import { ElMessage } from "element-plus";
import { checkSession } from "@/store/cache";
import { useSharedStore } from "@/store/sharedata";

const isCollapse = ref(false);
const router = useRouter();
const logo = ref("");
const curPath = ref("");
const chatListHeight = ref(0); // 聊天列表框高度
const title = ref("");
const avatarImg = ref("/images/avatar/default.jpg");
const store = useSharedStore();
const userInfo = store.getUserInfo;
const routerViewKey = ref(0);
const chatList = ref([]);
const chatListOffset = ref(0);
const chatListLimit = ref(20);
const loading = ref(false);
const finished = ref(false);
const chatId = ref("");

// 初始化 ChatID
chatId.value = router.currentRoute.value.params.id || "";
curPath.value = router.currentRoute.value.path;

// 监听路由变化;
router.beforeEach((to, from, next) => {
  curPath.value = to.path;
  if (curPath.value.includes("/chat")) {
    chatId.value = to.params.id || "";
  }
  next();
});

const newChat = () => {
  chatId.value = "";
  router.push('/chat').then(() => {
    // 刷新 `routerViewKey` 触发视图重新渲染
    routerViewKey.value += 1;
  });
}
const myReport = () => {
  chatId.value = "";
  router.push('/report').then(() => {
    // 刷新 `routerViewKey` 触发视图重新渲染
    routerViewKey.value += 1;
  });
}
// 切换会话
const loadChat = (chat) => {
  if (chatId.value === chat.chat_id) {
    return;
  }
  chatId.value = chat.chat_id;
  router.push(`/chat/${chat.chat_id}`).then(() => {
    // 刷新 `routerViewKey` 触发视图重新渲染
    routerViewKey.value += 1;
  });
};
const getChatList = () => {
  if (loading.value || finished.value) return;
  loading.value = true;
  httpPost("/api/chat/list", {
    offset: chatListOffset.value,
    limit: chatListLimit.value
  })
      .then((res) => {
        const items = res.data ? res.data : [];
        if (chatListOffset.value === 0) {
          chatList.value = items;
        } else {
          chatList.value = [...chatList.value, ...items];
        }

        if (chatList.value.length < res.count) {
          chatListOffset.value = chatListOffset.value + chatListLimit.value;
        } else {
          finished.value = true
        }
        loading.value = false;
      })
      .catch((e) => {
        loading.value = false;
        finished.value = true
        ElMessage.error("加载对话列表失败：" + e.message);
      });
};
onMounted(() => {
  init();
  resizeElement();
  window.onresize = () => resizeElement()
});

const init = () => {
  checkSession()
    .then((user) => {
      store.isLogin = true;
      store.setUserInfo(user)
      getChatList()
    })
    .catch(() => { });
};
const resizeElement = function () {
  chatListHeight.value = window.innerHeight - 398;
};
const goLogin = () => {
  router.push("/login");
}
const logout = function () {
  httpPost("/api/logout")
    .then(() => {
      store.setUserInfo({})
      store.setIsLogin(false);
      router.push("/login");
    })
    .catch(() => {
      ElMessage.error("注销失败！");
      router.push("/login");
    });
};

const loginSuccess = () => {
  init();
  store.setShowLoginDialog(false);
  // 刷新组件
  routerViewKey.value += 1;
};
const refreshChatList = () => {
  chatListOffset.value = 0;
  chatList.value = [];
  finished.value = false;
  getChatList();
}
provide('refreshChatList', refreshChatList);
</script>

<style lang="stylus" scoped>
@import "@/assets/css/custom-scroll.styl"
@import "@/assets/css/home.styl"
</style>
