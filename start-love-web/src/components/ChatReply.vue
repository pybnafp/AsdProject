<template>
  <div class="chat-line chat-line-reply-chat">
    <div class="chat-line-inner">
      <div class="chat-item">
        <div class="chat-status">
          <span v-if="thinking">思考中...</span>
          <span v-else>已完成思考</span>
          <el-icon>
            <ArrowDown/>
          </el-icon>
        </div>
        <div class="loading" v-loading="loading"></div>
        <div class="content-wrapper" v-show="data.reasoning">
          <el-divider direction="vertical" style="height: auto"/>
          <div class="content reasoning" v-html="md.render(processContent(data.reasoning))"></div>
        </div>
        <div class="content-wrapper">
          <div class="content" v-html="md.render(processContent(data.completion))"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import {ArrowDown} from "@element-plus/icons-vue";
import {processContent} from "@/utils/libs";
import hl from "highlight.js";
import emoji from "markdown-it-emoji";
import mathjaxPlugin from "markdown-it-mathjax3";
import MarkdownIt from "markdown-it";

// eslint-disable-next-line no-undef,no-unused-vars
const props = defineProps({
  data: {
    type: Object,
    default: {
      message_id: "",
      chat_id: "",
      user_id: 0,
      prompt: "",
      completion: "",
      files: null,
      created_at: "",
      reasoning: "",
    },
  },
  thinking: {
    type: Boolean,
    default: true,
  },
  loading: {
    type: Boolean,
    default: false,
  }
});

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
      const preCode = hl.highlight(str, {language: lang}).value;
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


</script>

<style lang="stylus">
@import '@/assets/css/markdown/vue.css';
.chat-line-reply-chat {
  justify-content: center;
  padding 20px;

  .chat-line-inner {
    display flex;
    width 100%;
    max-width 1000px;
    flex-flow row

    .chat-item {
      position: relative;
      padding: 2px 24px;
      overflow: hidden;

      .chat-status {
        padding: 1rem 0;
        font-size: 14px;
        line-height: 24px;
        color: #9E9E9E;

        span {
          margin-right 9px;
        }
      }
      .loading {
        margin-top 10px;
      }
      .content-wrapper {
        display flex
        .el-divider--vertical {
          border-left: 2px var(--el-border-color) var(--el-border-style);
          height: auto;
          width: 10px;
          margin-bottom: 15px;
          margin-left: 0;
        }
        .reasoning {
          font-size: 14px;
          line-height: 24px;
          color: #9E9E9E
          margin-bottom 15px;
        }
        .content {
          min-height 20px;
          word-break break-word;


          font-weight: normal;
          overflow auto;

          img {
            max-width: 600px;
            border-radius: 10px;
          }

          p {
            line-height 1.5

            code {
              color: var(--code-text-color);
              font-weight bold
              font-family: var(--font-family);
              background-color: var(--code-bg-color);
              border-radius: 4px;
              padding: .2rem .4rem;
            }
          }

          p:last-child {
            margin-bottom: 0
          }

          p:first-child {
            margin-top 0
          }

          .code-container {
            position relative
            display flex

            .hljs {
              border-radius 10px
              width 100%
            }

            .copy-code-btn {
              position: absolute;
              right 10px
              top 10px
              cursor pointer
              font-size 12px
              color #c1c1c1

              &:hover {
                color #20a0ff
              }
            }

          }

          .lang-name {
            position absolute;
            right 10px
            bottom 20px
            padding 2px 6px 4px 6px
            background-color #444444
            border-radius 10px
            color #00e0e0
          }


          // 设置表格边框

          table {
            width 100%
            margin-bottom 1rem
            border-collapse collapse;
            border 1px solid #dee2e6;
            background-color: var(--chat-content-bg);
            color: var(--theme-text-color-primary);

            thead {
              th {
                border 1px solid #dee2e6
                vertical-align: bottom
                border-bottom: 2px solid #dee2e6
                padding 10px
              }
            }

            td {
              border 1px solid #dee2e6
              padding 10px
            }
          }

          // 代码快

          blockquote {
            margin 0
            background-color: #ebfffe;
            padding: 0.8rem 1.5rem;
            border-left: 0.5rem solid;
            border-color: #026863;
            color: #2c3e50;
          }
        }

      }
    }
  }

}
</style>
