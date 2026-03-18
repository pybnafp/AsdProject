<template>
  <el-container class="file-select-box">
    <a class="file-upload-img" @click="fetchReports">
      <img src="@/assets/img/icon-attachment.svg" alt="上传报告">
    </a>
    <van-popup v-model:show="show" position="bottom" round closeable>
      <div class="title">从我的报告中添加</div>
      <div class="filter-box">
        <div class="filter-item" :class="item.value === selectedFilter ? 'active' : ''" v-for="item in filterItems"
             :key="item.value" @click="onSelectFilter(item.value)">
          {{ item.label }}
        </div>
      </div>
      <div class="report-list">
        <van-list
            v-model:error="error"
            v-model:loading="loading"
            :finished="finished"
            error-text="请求失败，点击重新加载"
            finished-text="没有更多了"
            @load="onLoad"
        >
          <div class="card" v-for="report in reports" :key="report.report_id">
            <div class="header">
              <span class="header-left">{{ report.name }}</span>
              <el-button round plain type="primary" size="small" @click="selectReport(report)">
                选择
              </el-button>
            </div>
            <div class="content">
              <div class="content-row">
                <span>文字报告</span>
                <span>{{ report.report_file_name }}</span>
              </div>
              <div class="content-row">
                <span>上传时间</span>
                <span>{{ report.created_at }}</span>
              </div>
              <div class="content-row">
                <span>报告日期</span>
                <span>{{ report.report_date }}</span>
              </div>
            </div>
          </div>
        </van-list>
      </div>
      <el-upload
          class="upload-box"
          drag
          :auto-upload="true"
          :show-file-list="false"
          :http-request="afterRead"
          accept=".doc,.docx,.jpg,.png,.jpeg,.xls,.xlsx,.ppt,.pptx,.pdf"
      >
        <img src="@/assets/img/icon-upload.svg" alt="">
        <div class="el-upload__text">
          <p style="font-size: 13px; line-height:18px;color: #000000">上传文件</p>
          <p style="font-size: 10px; line-height:12px;color: #6E6E6E">请将文件拖入该区域内</p>
        </div>
      </el-upload>
    </van-popup>
  </el-container>
</template>

<script setup>
import {onMounted, ref} from "vue";
import {ElMessage} from "element-plus";
import {httpPost} from "@/utils/http";
import {showFailToast} from "vant";

const emits = defineEmits(["selectReport", "uploadFile"]);
const show = ref(false);
const scrollbarRef = ref(null);
const loading = ref(false)
const finished = ref(false)
const error = ref(false)
const offset = ref(0);
const limit = ref(20);
const reports = ref([]);
const filterItems = ref([
  {"label": "2023年", "value": 2023},
  {"label": "2024年", "value": 2024},
  {"label": "2025年", "value": 2025},
]);
const selectedFilter = ref(2025);
onMounted(() => {
});

const onSelectFilter = (year) => {
  selectedFilter.value = year;
  offset.value = 0;
  loading.value = false;
  finished.value = false;
  onLoad();
}
const selectReport = (report) => {
  show.value = false;
  emits("selectReport", report);
}
const insertFile = (file) => {
  show.value = false;
  emits("uploadFile", file);
};
const fetchReports = () => {
  // 临时开启
  show.value = true;
  onLoad();
}
const onLoad = () => {
  httpPost("/api/reports/list", {
    offset: offset.value,
    limit: limit.value,
    year: selectedFilter.value,
  }).then((res) => {
    loading.value = false;
    if (res.data) {
      reports.value = res.data;
      const items = res.data;

      if (offset.value === 0) {
        reports.value = items;
      } else {
        reports.value = [...reports.value, ...items];
      }

      if (reports.value.length < res.count) {
        offset.value = offset.value + limit.value;
      } else {
        finished.value = true
      }
    } else {
      if (offset.value === 0) {
        reports.value = [];
      }
      finished.value = true
    }
  }).catch(() => {
    loading.value = false;
    error.value = true
    showFailToast("加载报告列表失败")
  })
};

// el-scrollbar 滚动回调
const onScroll = (options) => {
  const wrapRef = scrollbarRef.value.wrapRef;
  scrollbarRef.value.moveY = (wrapRef.scrollTop * 100) / wrapRef.clientHeight;
  scrollbarRef.value.moveX = (wrapRef.scrollLeft * 100) / wrapRef.clientWidth;
  const poor = wrapRef.scrollHeight - wrapRef.clientHeight;
  // 判断滚动到底部 自动加载数据
  if (options.scrollTop + 2 >= poor) {
    fetchReports();
  }
};

const afterRead = (file) => {
  const formData = new FormData();
  formData.append("file", file.file, file.name);
  // 执行上传操作
  httpPost("/api/files/upload", formData)
      .then((res) => {
        ElMessage.success({message: "上传成功", duration: 500});
        insertFile(res.data);
      })
      .catch((e) => {
        ElMessage.error("上传失败:" + e.message);
      });
};
</script>

<style lang="stylus">

.file-select-box {
  .file-upload-img {
    .img {
      width: 20px;
      height: 20px;
    }
  }
  .title {
    font-size 16px;
    line-height 50px;
    text-align left;
    padding: 0 16px;
  }
  .filter-box {
    margin-top 10px;
    display: flex;
    gap: 16px;
    padding: 0 16px;
    .filter-item {
      padding: 0 16px;
      border-radius: 372px;
      background: #FDFDFD;
      border: 0.5px solid #F0F1F2;
      font-size: 14px;
      color: #6D6A71;
      height: 38px;
      line-height: 38px;

      &.active {
        background: rgba(93, 25, 210, 0.1);
        color: #5D19D2;
      }
    }
  }

  .report-list {
    margin 10px 15px;
    max-height 300px;
    overflow-y scroll;
    border-radius: 12px;
    background: rgba(93, 25, 210, 0.1);
    scrollbar-width: none; /* 隐藏滚动条（Firefox） */
    .van-list {
      .card {
        border-radius: 12px;
        box-shadow: 0px 2px 61px 0px #F0EDF6;
        margin 2px;
        padding 14px 18px;
        background #ffffff;

        .header {
          line-height: 18px;
          display flex;
          justify-content space-between;

          .header-left {
            font-size: 14px;
          }
        }

        .content {
          .content-row {
            display flex;
            justify-content start;
            gap 16px;

            padding 6px 0
            font-size: 12px;
            color: #6E6E6E;
            span:first-child {
              flex-shrink: 0;
            }
          }
        }
      }
    }
  }
  .my-report-table {
    border-radius: 16px;
    box-sizing: border-box;
    border: 2px solid rgba(93, 25, 210, 0.1);
    padding: 10px;
    margin: 25px 0;
  }

  .upload-box {
    margin 12px 16px 40px 16px;
    .el-upload-dragger {
      width: 100%;
      height: 88px;
      border-radius: 16px;
      background: rgba(93, 25, 210, 0.05);
      box-sizing: border-box;
      border: 1px dashed #5D19D2;
      display flex;
      justify-content center;
      align-items center;
      img {
        width 50px;
        height 50px;
      }
      .el-upload__text {
        text-align left
      }
    }
  }
}
</style>
