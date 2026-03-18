<template>
  <el-container class="file-select-box">
    <a class="file-upload-img" @click="fetchReports">
      <img src="@/assets/img/icon-attachment.svg" alt="上传报告">
    </a>
    <el-dialog class="file-list-dialog" v-model="show" :close-on-click-modal="true" :show-close="true" :width="800"
               title="从我的报告中添加" :center="false">
      <template #header="{titleId, titleClass }">
        <div class="my-header">
          <span :id="titleId" :class="titleClass">从我的报告中添加</span>
        </div>
      </template>
      <div class="filter-box">
        <div class="filter-item" :class="item.value === selectedFilter ? 'active' : ''" v-for="item in filterItems"
             :key="item.value" @click="onSelectFilter(item.value)">
          {{ item.label }}
        </div>
      </div>
      <div class="my-report-table">
        <el-table
            ref="multipleTableRef"
            :data="reportTableData"
            row-key="id"
            style="width: 100%"
            max-height="238"
            stripe
        >
          <el-table-column property="report_id" label="ID" width="100"/>
          <el-table-column property="name" label="姓名" />
          <el-table-column property="report_file_name" label="文字报告"/>
          <el-table-column label="上传时间">
            <template #default="scope">{{ scope.row.created_at }}</template>
          </el-table-column>
          <el-table-column label="报告日期">
            <template #default="scope">{{ scope.row.report_date }}</template>
          </el-table-column>
          <el-table-column label="操作" width="65">
            <template #default="scope">
              <el-button round plain type="primary" size="small" @click="selectReport(scope.row)">
                选择
              </el-button>
            </template>
          </el-table-column>
        </el-table>
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
          <p style="font-size: 18px; line-height:42px;color: #000000">上传文件</p>
          <p>请将文件拖入该区域内</p>
        </div>
      </el-upload>
    </el-dialog>
  </el-container>
</template>

<script setup>
import {onMounted, ref} from "vue";
import {ElMessage} from "element-plus";
import {httpPost} from "@/utils/http";
import {useSharedStore} from "@/store/sharedata";
import {showMessageError} from "@/utils/dialog";

const props = defineProps({});
const emits = defineEmits(["selectReport", "uploadFile"]);
const store = useSharedStore();
const show = ref(false);
const scrollbarRef = ref(null);
const offset = ref(0);
const limit = ref(20);
const reportList = ref([]);
const filterItems = ref([
  {"label": "2023年", "value": 2023},
  {"label": "2024年", "value": 2024},
  {"label": "2025年", "value": 2025},
]);
const selectedFilter = ref(2025);

const multipleTableRef = ref()

const reportTableData = ref([]);
onMounted(() => {
});

const onSelectFilter = (year) => {
  selectedFilter.value = year;
  offset.value = 0;
  fetchReports()
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
  httpPost("/api/reports/list", {
    offset: offset.value,
    limit: limit.value,
    year: selectedFilter.value,
  })
      .then((res) => {
        if (!res.data) {
          return;
        }
        const items = res.data;

        if (offset.value === 0) {
          reportList.value = items;
        } else {
          reportList.value = [...reportList.value, ...items];
        }
        reportTableData.value = reportList.value
        if (reportList.value.length < res.data.count) {
          offset.value = offset.value + 1;
        }
      })
      .catch((e) => {
        showMessageError("获取报告列表失败：" + e.message);
      });
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

  .el-dialog {

    .el-dialog__body {
      //padding 0
      overflow hidden

      .filter-box {

        /* 自动布局 */
        display: flex;
        gap: 16px;

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

      .my-report-table {
        border-radius: 16px;
        box-sizing: border-box;
        border: 2px solid rgba(93, 25, 210, 0.1);
        padding: 10px;
        margin: 25px 0;
      }

      .upload-box {
        margin 0;
        .el-upload-dragger {
          width: 765px;
          height: 255px;
          border-radius: 16px;
          background: rgba(93, 25, 210, 0.05);
          box-sizing: border-box;
          border: 1px dashed #5D19D2;
          display flex;
          flex-direction column;
          justify-content center;
          align-items center;

          img {
            width 50px;
            height 50px;
            margin-bottom 14px
          }
          .el-upload__text {
            text-align center;
          }
        }
      }
    }
  }
}

.my-header {
  font-size: 18px;
  line-height: 18px;
  text-align left
}
</style>
