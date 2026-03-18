<template>
  <div>
    <div class="my-report-table">
      <el-table
          ref="multipleTableRef"
          :data="reports"
          row-key="id"
          style="width: 100%"
          :max-height="tableHeight"
          stripe
          @selection-change="handleSelectionChange"
      >
        <el-table-column property="report_id" label="ID"/>
        <el-table-column property="name" label="姓名"/>
        <el-table-column property="original_file_name" label="原数据"/>
        <el-table-column property="report_file_name" label="文字报告">
          <template #default="scope">
            <el-link :href="scope.row.report_file_url" target="_blank" :underline="false">{{ scope.row.report_file_name}}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="上传时间">
          <template #default="scope">{{ scope.row.created_at }}</template>
        </el-table-column>
        <el-table-column label="报告日期">
          <template #default="scope">{{ scope.row.report_date }}</template>
        </el-table-column>
        <el-table-column property="status" label="状态">
          <template #default="scope">
            <el-text :type="formatReportStatus(scope.row.status).type">{{
                formatReportStatus(scope.row.status).label
              }}
            </el-text>
          </template>
        </el-table-column>
        <template #empty>
          <div class="empty-data-box">
            <img src="@/assets/img/icon-empty.svg" alt="">
            <p>您还没有单项报告，去看看别的吧</p>
          </div>
        </template>
      </el-table>
    </div>
    <div class="table-pagination">
      <el-pagination background layout="prev, pager, next, jumper" :total="count" :size="limit"
                     @size-change="handleSizeChange"/>
    </div>
  </div>
</template>
<script setup>
import {onMounted, ref, watch} from "vue";
import {httpPost} from "@/utils/http";
import {showMessageError} from "@/utils/dialog";

// eslint-disable-next-line no-undef
const props = defineProps({
  type: Number,
});
const offset = ref(0)
const limit = ref(10)
const count = ref(0)
watch(
    () => props.type,
    (newValue, oldValue) => {
      if (newValue !== oldValue) {
        offset.value = 0;
        reports.value = [];
      }
      fetchReports()
    }
);

const tableHeight = ref(0);
const resizeElement = function () {
  if (props.type === 1) { //综合报告
    tableHeight.value = window.innerHeight - 282;
  } else {
    tableHeight.value = window.innerHeight - 320;
  }
};


const multipleTableRef = ref()
const multipleSelection = ref([])

const reports = ref([]);
onMounted(() => {
  resizeElement();
  fetchReports();
  if (reports.value) {
    document.getElementsByClassName("el-pagination__goto")[0].childNodes[0].nodeValue = "前往";
    document.getElementsByClassName("el-pagination__goto")[1].childNodes[0].nodeValue = "前往";
  }

});
const handleSelectionChange = (val) => {
  multipleSelection.value = val
}
const handleSizeChange = (page) => {
  offset.value = (page - 1) * limit.value;
  fetchReports()
}
const fetchReports = () => {
  httpPost("/api/reports/list", {offset: offset.value, limit: limit.value, type: props.type})
      .then((res) => {
        if (!res.data) {
          return;
        }
        count.value = res.count;
        reports.value = res.data;
      })
      .catch((e) => {
        showMessageError("获取报告列表失败：" + e.message);
      });
};
const formatReportStatus = (status) => {
  switch (status) {
    case 1:
      return {"type": "warning", "label": "待审核"};
    case 2:
      return {"type": "success", "label": "已审核"};
    default :
      return {"type": "success", "label": "已审核"};
  }
}
</script>
<style scoped lang="stylus">
.my-report-table {
  border-radius: 16px;
  box-sizing: border-box;
  border: 2px solid rgba(93, 25, 210, 0.1);
  padding: 10px;
  margin: 25px 0;
}

.table-pagination {
  display flex;
  justify-content center;
}

.empty-data-box {
  height: 350px;
  display flex;
  flex-direction column;
  justify-content center;
  align-items center

  p {
    font-size: 12px;
    line-height: 24px;
    color: #26292B;
    margin-top 26px;
  }
}
</style>