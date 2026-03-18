<template>
  <div class="chat-line chat-line-prompt-chat">
    <div class="chat-line-inner">
      <div class="chat-item">
        <div class="content-wrapper">
          <div class="content" v-html="prompt"></div>
        </div>
        <div v-if="files && files.length > 0" class="chat-file-list">
          <div v-for="file in files" :key="file.file_id" >
            <div class="file-item">
              <img :src="GetFileIcon(file.file_type)" alt="">
              <div class="body">
                <div class="title">
                  {{ substr(file.file_name, 30) }}
                </div>
                <div class="info">
                  <span>{{ GetFileType(file.file_type) }}·</span>
                  <span>{{ FormatFileSize(file.file_size) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div v-if="reports && reports.length > 0" class="chat-file-list">
          <div v-for="report in reports" :key="report.report_id" >
            <div class="file-item">
              <img :src="GetFileIcon('.docx')" alt="">
              <div class="body">
                <div class="title">
                  {{ substr(report.report_file_name, 30) }}
                </div>
                <div class="info">
                  <span>{{ GetFileType('.docx') }}·</span>
                  <span>{{ FormatFileSize(report.report_file_size) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import {onMounted, ref} from "vue";
import hl from "highlight.js";
import {processPrompt, substr} from "@/utils/libs";
import {FormatFileSize, GetFileIcon, GetFileType} from "@/store/system";
import emoji from "markdown-it-emoji";
import mathjaxPlugin from "markdown-it-mathjax3";
import MarkdownIt from "markdown-it";

const md = new MarkdownIt({
  breaks: true,
  html: true,
  linkify: true,
  typographer: true,
  highlight: function (str, lang) {
    const codeIndex = parseInt(Date.now()) + Math.floor(Math.random() * 10000000);
    // 显示复制代码按钮
    const copyBtn = `<span class="copy-code-btn" data-clipboard-action="copy" data-clipboard-target="#copy-target-${codeIndex}">复制</span>
<textarea style="position: absolute;top: -9999px;left: -9999px;z-index: -9999;" id="copy-target-${codeIndex}">${str.replace(
        /<\/textarea>/g,
        "&lt;/textarea>"
    )}</textarea>`;
    if (lang && hl.getLanguage(lang)) {
      const langHtml = `<span class="lang-name">${lang}</span>`;
      // 处理代码高亮
      const preCode = hl.highlight(lang, str, true).value;
      // 将代码包裹在 pre 中
      return `<pre class="code-container"><code class="language-${lang} hljs">${preCode}</code>${copyBtn} ${langHtml}</pre>`;
    }

    // 处理代码高亮
    const preCode = md.utils.escapeHtml(str);
    // 将代码包裹在 pre 中
    return `<pre class="code-container"><code class="language-${lang} hljs">${preCode}</code>${copyBtn}</pre>`;
  },
});
md.use(mathjaxPlugin);
md.use(emoji);
const props = defineProps({
  data: {
    type: Object,
    default: {
      message_id: "",
      chat_id: "",
      user_id: 0,
      prompt: "",
      completion: "",
      reasoning: "",
      files: [],
      reports: [],
      created_at: "",
    },
  },
});
const prompt = ref(processPrompt(props.data.prompt));
const files = ref(props.data.files ? props.data.files : []);
const reports = ref(props.data.reports ? props.data.reports : []);
onMounted(() => {
  processFiles();
});

const processFiles = () => {
  if (!props.data.prompt) {
    return;
  }
  prompt.value = md.render(prompt.value.trim())
}
</script>

<style scoped lang="stylus">
@import '@/assets/css/markdown/vue.css';
.chat-line-prompt-chat {
  justify-content: center;
  width 100%
  padding-bottom: 1.5rem;
  padding-top: 1.5rem;

  .chat-line-inner {
    display flex;
    width 100%;
    max-width 1000px;
    padding 0 20px;
    flex-flow row-reverse

    .chat-item {
      overflow: hidden;
      max-width 60%;
      background: #F8F5FE;
      box-sizing: border-box;
      border: 1px solid #E5E5E5;
      border-radius: 14px 14px 2px 14px;
      padding: 2px 24px;

      .content-wrapper {
        display flex
        flex-flow row-reverse

        .content {
          word-break break-word;
          font-size: 14px;
          overflow: auto;
        }
      }

      .chat-file-list {
        display flex
        flex-flow row

        .file-item {
          position relative
          display flex
          flex-flow row
          border-radius 12px
          background-color: #F8F5FE;
          margin-right 10px;
          padding-bottom 10px;


          img {
            width 44px
            height 44px
          }

          .body {
            margin-left 8px
            text-align left

            .title {
              font-size 16px
            }

            .info {
              font-size: 12px;
              color #6E6E6E;
              margin-top 3px;
            }
          }
        }
      }
    }
  }
}

</style>
