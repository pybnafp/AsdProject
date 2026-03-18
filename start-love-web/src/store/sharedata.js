import { defineStore } from "pinia";
import Storage from "good-storage";

export const useSharedStore = defineStore("shared", {
  state: () => ({
    showLoginDialog: false,
    theme: Storage.get("theme", "light"),
    isLogin: false,
    userInfo: Storage.get("userInfo", null),
  }),
  getters: {
    getUserInfo: (state) => state.userInfo,
  },
  actions: {
    setShowLoginDialog(value) {
      this.showLoginDialog = value;
    },
    setTheme(theme) {
      this.theme = theme;
      document.documentElement.setAttribute("data-theme", theme); // 设置 HTML 的 data-theme 属性
      Storage.set("theme", theme);
    },
    setIsLogin(value) {
      this.isLogin = value;
    },
    setUserInfo(value) {
      this.userInfo = value;
      Storage.set("userInfo", value)
    },
  },
});
