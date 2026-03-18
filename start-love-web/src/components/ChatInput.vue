<template>
  <div class="input-box">
    <div class="input-box-inner">
      <div class="input-body">
        <div ref="textHeightRef" class="hide-div">{{ prompt }}</div>
        <div class="input-border">
          <div class="input-inner">
            <div class="file-list" v-if="files.length > 0 || reports.length > 0">
              <FileList :files="files" @remove-file="removeFile"/>
              <ReportList :reports="reports" @remove-report="removeReport"/>
            </div>
            <textarea ref="inputRef" class="prompt-input" :rows="row" v-model="prompt" @keydown="onInput"
                      @input="onInput" placeholder="输入您的问题..." autofocus>
                        </textarea>
          </div>
          <div class="flex-between">
            <div class="flex little-btns">
              <span class="tool-item-btn" :class="isGuide ? 'active' : ''" @click="isGuide = !isGuide">
                <i class="iconfont icon-zhinanzhen"></i>
                <span>医疗指南</span>
              </span>
              <span class="tool-item-btn" :class="isInnovate ? 'active' : ''" @click="isInnovate = !isInnovate">
                <i class="iconfont icon-chuangxinyanjiu"></i>
                <span>研究创新</span>
              </span>
              <span class="tool-item-btn">
                <MobileFileSelect v-if="isMobileV2()" @selectReport="insertReport" @uploadFile="insertFile"/>
                <FileSelect v-else @selectReport="insertReport" @uploadFile="insertFile"/>
              </span>
            </div>
            <div class="flex">
              <span class="send-btn">
                <!-- showStopGenerate -->
                <el-button type="info" v-if="showStopGenerate" @click="stopGenerate" plain>
                  <el-icon>
                  </el-icon>
                </el-button>
                <el-button @click="send" :circle="true" color="#5D19D2" v-else>
                  <el-tooltip class="box-item" effect="dark" content="发送">
                    <i class="iconfont icon-fasong" style="color: #ffffff; font-size: 24px"></i>
                  </el-tooltip>
                </el-button>
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
    <!-- end input box -->
    <div></div>
  </div>
</template>
<script setup>
import {onMounted, ref, inject} from "vue";
import FileSelect from "@/components/FileSelect.vue";
import MobileFileSelect from "@/components/mobile/FileSelect.vue";
import FileList from "@/components/FileList.vue";
import ReportList from "@/components/ReportList.vue";
import { httpPost } from "@/utils/http";
import {getValuesByProperty, isMobileV2, removeArrayItem} from "@/utils/libs";
import {showMessageError} from "@/utils/dialog";
import {ElMessage} from "element-plus";
import { showNotify } from 'vant';


const prompt = ref("");
const files = ref([
  // {
  //   "id": 1,
  //   "name": "冯子园的报告",
  //   "ext": "pdf",
  //   "size": "10240"
  // },
  // {
  //   "id": 2,
  //   "name": "冯子园的报告",
  //   "ext": "pdf",
  //   "size": "10240"
  // },
]);
const reports = ref([]);
const row = ref(1);
const inputRef = ref(null);
const isGuide = ref(false);
const isInnovate = ref(false);
const showStopGenerate = ref(false); // 停止生成
const canSend = ref(true);
const textHeightRef = ref(null);
const emits = defineEmits(["send"]);

const disableInput = (force) => {
  canSend.value = false;
  showStopGenerate.value = !force;
};

const enableInput = () => {
  canSend.value = true;
  showStopGenerate.value = false;
};

const stopGenerate = function () {
  showStopGenerate.value = false;
  httpPost("/api/chat/stop_stream", {
    chat_id: "",
    message_id: "",
  }).then(() => {
    enableInput();
  });
};

const onInput = (e) => {
  // 根据输入的内容自动计算输入框的行数
  const lineHeight = parseFloat(window.getComputedStyle(inputRef.value).lineHeight);
  textHeightRef.value.style.width = inputRef.value.clientWidth + "px"; // 设定宽度和 textarea 相同
  const lines = Math.floor(textHeightRef.value.clientHeight / lineHeight);
  inputRef.value.scrollTo(0, inputRef.value.scrollHeight);
  if (prompt.value.length < 10) {
    row.value = 1;
  } else if (lines <= 7) {
    row.value = lines;
  } else {
    row.value = 7;
  }

  // 输入回车自动提交
  if (e.keyCode === 13) {
    if (e.ctrlKey) {
      // Ctrl + Enter 换行
      prompt.value += "\n";
      return;
    }
    e.preventDefault()
    send();
  }
};

const insertReport = (report) => {
  reports.value.push(report);
};
const removeReport = (report) => {
  reports.value = removeArrayItem(reports.value, report, (v1, v2) => v1.report_id === v2.report_id);
};
// 插入文件
const insertFile = (file) => {
  files.value.push(file);
};
const removeFile = (file) => {
  files.value = removeArrayItem(files.value, file, (v1, v2) => v1.file_id === v2.file_id);
};


const send = function () {
  if (canSend.value === false) {
    ElMessage.warning("AI 正在作答中，请稍后...");
    return;
  }
  // disableInput();
  if (prompt.value.trim().length === 0 || canSend.value === false) {
    if (isMobileV2()) {
      showNotify({ message: '请输入要发送的消息！' });
    } else {
      showMessageError("请输入要发送的消息！");

    }
    return false;
  }

  let chatItem = {
    prompt: prompt.value,
    chat_id: "",
    file_ids: getValuesByProperty(files.value, 'file_id'),
    report_ids: getValuesByProperty(reports.value, 'report_id'),
    enable_guidelines: isGuide.value,
    enable_researches: isInnovate.value,
  };
  emits("send", chatItem);
  prompt.value = "";
  files.value = [];
  reports.value = [];
}
</script>
<style scoped lang="stylus">
.input-box {
  width 100%
  max-width 800px;
  margin auto;

  .input-box-inner {
    display flex
    justify-content: center;
    align-items: center;

    .input-body {
      width 100%
      margin: 0;
      border: none;
      padding: 10px 0;
      display flex
      justify-content center
      position relative

      .hide-div {
        white-space: pre-wrap; /* 保持文本换行 */
        visibility: hidden; /* 隐藏 div */
        position: absolute; /* 脱离文档流 */
        line-height: 24px
        font-size 14px
        word-wrap: break-word; /* 允许单词换行 */
        overflow-wrap: break-word; /* 允许长单词换行，适用于现代浏览器 */
      }

      .input-border {
        // display flex
        width 100%
        overflow hidden
        border: 1px solid #F8F5FE;
        border-radius 24px
        padding 12px
        box-shadow: 0px 17px 40px 0px #F0EDF6;
        background: #FFFFFF;


        &:hover {
          border-color var(--theme-border-hover)
        }

        .input-inner {
          display flex
          flex-flow column
          width 100%

          .file-list {
            padding-bottom 10px
          }

          .prompt-input::-webkit-scrollbar {
            width: 0;
            height: 0;
          }

          .prompt-input {
            min-height: 58px;
            width 100%
            line-height: 24px
            border none
            font-size 14px
            background none
            resize: none
            white-space: pre-wrap; /* 保持文本换行 */
            word-wrap: break-word; /* 允许单词换行 */
            overflow-wrap: break-word; /* 允许长单词换行，适用于现代浏览器 */
          }
        }


        .send-btn {

          margin-left: 10px

          .el-button {
            width: 48px;
            height: 48px;
            border-radius:24px;
          }
        }

        .little-btns {
          display: flex;
          justify-content: flex-end;
          align-items: center;
          gap: 8px;

          .tool-item-btn {
            border-radius: 120px;
            box-sizing: border-box;
            border: 0.5px solid rgba(93, 25, 210, 0.1);
            cursor: pointer;
            padding: 6px 10px;
            font-size: 14px;
            letter-spacing: 0.16em;
            display: flex;
            align-items : center;
            height: 32px;

            i {
              font-size: 20px;
              margin-right: 4px;
            }

            &.active {
              background: rgba(93, 25, 210, 0.1);
              border: 0.5px solid #F0F1F2;
              color: var(--text-color)
            }
          }

        }

        .add-new {
          .el-icon {
            font-size: 20px;
            color: #754ff6;
          }
          cursor: pointer
        }

      }
    }
  }

}
</style>