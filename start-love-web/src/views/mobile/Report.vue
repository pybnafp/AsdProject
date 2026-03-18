<template>
  <div class="app-background">
    <div class="report-page">
      <van-tabs v-model="activeTab" @click-tab="onTabClick" animated>
        <van-tab title="单项报告" name="singleReport">
          <div class="filter-box">
            <div class="filter-item" :class="item.type === selectedFilter ? 'active' : ''" v-for="item in filterItems"
                 :key="item.type" @click="onSelectFilter(item)">
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
                 <span class="header-left">{{report.name}}</span>
                 <span class="header-right">
                    <el-text :type="formatReportStatus(report.status).type">
                      {{ formatReportStatus(report.status).label }}
                    </el-text>
                 </span>
               </div>
               <div class="content">
                 <div class="content-row">
                   <span>原数据</span>
                   <span>{{report.original_file_name}}</span>
                 </div>
                 <div class="content-row">
                   <span>文字报告</span>
                   <span>{{report.report_file_name}}</span>
                 </div>
                 <div class="content-row">
                   <span>上传时间</span>
                   <span>{{report.created_at}}</span>
                 </div>
                 <div class="content-row">
                   <span>报告日期</span>
                   <span>{{report.report_date}}</span>
                 </div>
               </div>
             </div>
            </van-list>
          </div>
        </van-tab>
        <van-tab title="综合报告" name="generalReport">
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
                  <span class="header-left">{{report.user_id}}</span>
                  <span class="header-right">
                    <el-text :type="formatReportStatus(report.status).type">
                      {{ formatReportStatus(report.status).label }}
                    </el-text>
                  </span>
                </div>
                <div class="content">
                  <div class="content-row">
                    <span>原数据</span>
                    <span>{{report.original_file_name}}</span>
                  </div>
                  <div class="content-row">
                    <span>文字报告</span>
                    <span>{{report.report_file_name}}</span>
                  </div>
                  <div class="content-row">
                    <span>上传时间</span>
                    <span>{{report.created_at}}</span>
                  </div>
                  <div class="content-row">
                    <span>报告日期</span>
                    <span>{{report.report_date}}</span>
                  </div>
                </div>
              </div>
            </van-list>
          </div>
        </van-tab>
      </van-tabs>
    </div>
  </div>
</template>
<script setup>
import {ref} from "vue";
import {httpPost} from "@/utils/http";
import {showFailToast} from "vant";
import {checkSession} from "@/store/cache";

const activeTab = ref('singleReport')
const reports = ref([])
const loading = ref(false)
const finished = ref(false)
const error = ref(false)
const offset = ref(0)
const limit = ref(10)
const filterItems = ref([
  {"label": "问卷量表", "type": 2},
  {"label": "面容数据", "type": 3},
  {"label": "眼动数据", "type": 4},
  {"label": "脑影像数据", "type": 5},
  {"label": "行为视频", "type": 6},
  {"label": "基因数据", "type": 7},
]);
const generalReportType = ref(1);
const selectedFilter = ref(filterItems.value[0].type);
const currentType = ref(filterItems.value[0].type);

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

const onSelectFilter = (item) => {
  selectedFilter.value = item.type;
  currentType.value = item.type;
  offset.value = 0;
  loading.value = false;
  finished.value = false;
}
const onTabClick = (tab) => {
  if (tab.name === "generalReport") {
    currentType.value = generalReportType.value;
  } else {
    selectedFilter.value = filterItems.value[0].type;
    currentType.value = filterItems.value[0].type;
  }
  offset.value = 0;
  onLoad();
}
const onLoad = () => {
  httpPost("/api/reports/list", {
    offset: offset.value,
    limit: limit.value,
    type: currentType.value,
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
</script>
<style lang="stylus" scoped>
@import "@/assets/css/mobile/report.styl"
</style>