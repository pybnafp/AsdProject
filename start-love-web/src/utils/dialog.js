/**
 * Util lib functions
 */
import {closeToast, showConfirmDialog, showFailToast, showLoadingToast, showSuccessToast, showToast} from "vant";
import {isMobileV2} from "@/utils/libs";
import {ElMessage} from "element-plus";

export function showLoginDialog(router) {
  showConfirmDialog({
    title: "登录",
    message: "此操作需要登录才能进行，前往登录？",
  })
      .then(() => {
        router.push("/login");
      })
      .catch(() => {
        // on cancel
    });
}

export function showMessageOK(message) {
  if (isMobileV2()) {
    showSuccessToast(message);
  } else {
    ElMessage.success(message);
  }
}

export function showMessageInfo(message) {
  if (isMobileV2()) {
    showToast(message);
  } else {
    ElMessage.info(message);
  }
}

export function showMessageError(message) {
  if (isMobileV2()) {
    showFailToast({message: message, duration: 2000});
  } else {
    ElMessage.error(message);
  }
}

export function showLoading(message = "正在处理...") {
  showLoadingToast({message: message, forbidClick: true, duration: 2000});
}

export function closeLoading() {
  closeToast();
}
