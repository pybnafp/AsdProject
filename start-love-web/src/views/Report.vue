<template>
  <div class="report-page">
    <div class="page-header">我的报告</div>
    <el-divider />
    <div>
      <el-tabs v-model="activeName" @tab-click="handleTabClick">
        <el-tab-pane label="单项报告" name="first">
          <template #label>
            <span>·单项报告</span>
          </template>
          <div class="filter-box">
            <div class="filter-item" :class="item.type === selectedFilter ? 'active' : ''" v-for="item in filterItems"
                 :key="item.type" @click="onSelectFilter(item)">
              {{ item.label }}
            </div>
          </div>
          <ReportTable :type="selectedFilter"/>
        </el-tab-pane>
        <el-tab-pane label="综合报告" name="second">
          <template #label>
            <span>·综合报告</span>
          </template>
          <ReportTable :type="reportDefaultType"/>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>
<script setup>
import {onMounted, ref} from 'vue';
import ReportTable from "@/components/ReportTable.vue";

const activeName = ref('first');
const filterItems = ref([
  {"label": "问卷量表", "type": 2},
  {"label": "面容数据", "type": 3},
  {"label": "眼动数据", "type": 4},
  {"label": "脑影像数据", "type": 5},
  {"label": "行为视频", "type": 6},
  {"label": "基因数据", "type": 7},
]);
const reportDefaultType = ref(1);
const selectedFilter = ref(filterItems.value[0].type);

onMounted(() => {

});
const handleTabClick = (tab, event) => {
  console.log(tab, event)
}
const onSelectFilter = (item) => {
  selectedFilter.value = item.type;
}
</script>
<style lang="stylus">
@import "@/assets/css/report.styl"
</style>