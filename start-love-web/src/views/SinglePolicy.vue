<template>
  <div class="single-page">
    <el-container>
      <el-header>
        <header-bar></header-bar>
      </el-header>
      <el-main>
        <div id="container" :style="{ height: mainWinHeight + 'px' }">
          <div class="news-content">
            <div class="news-row" v-for="(news,index) in newsList" :key="index">
              <a :href="news.href" target="_blank">
                <el-row>
                  <el-col :span="20">
                    <span>{{ news.title }}</span>
                  </el-col>
                  <el-col :span="3">
                    <span>{{ news.date }}</span>
                  </el-col>
                  <el-col :span="1">
                    <img src="@/assets/img/icon-arrows-right.png" alt=""/>
                  </el-col>
                </el-row>
              </a>
            </div>
            <div class="pagination">
              <el-pagination background layout="prev, pager, next, jumper" :total="count" :size="limit"
                             @size-change="handleSizeChange"/>
            </div>
          </div>
        </div>
      </el-main>
    </el-container>
  </div>
</template>
<script setup>
import {onMounted, ref} from 'vue';
import HeaderBar from "@/components/HeaderBar.vue";

const mainWinHeight = ref(0); // 主窗口高度
const limit = ref(10)
const count = ref(10)
const newsList = ref([
  {
    title: "中国残联等七部门联合印发孤独症儿童关爱促进行动实施方案",
    date: "2024-07-26",
    href: "https://mp.weixin.qq.com/s/Ugs1q_fmK97uJZC6RZIN0g",
  },
  {
    title: "孤独症儿童上学支持性政策汇编",
    date: "2024-09-05",
    href: "https://mp.weixin.qq.com/s/_8PYHgxH8dUPeHwtANHn5w",
  },
  {
    title: "国家有关孤独症患者权益保障的法律法规和条例",
    date: "2024-03-18",
    href: "https://mp.weixin.qq.com/s/p0QlKI56xQmkNesp_CIWVw",
  },
  {
    title: "自闭症家庭迎来五大重磅支持政策",
    date: "2025-05-20",
    href: "https://mp.weixin.qq.com/s/JwFLsZ_nLinSYJSiop3TDA",
  },
  {
    title: "国家给的这笔钱，很多自闭症家长还没领！2025最全康复补贴政策来了",
    date: "2025-04-28",
    href: "https://mp.weixin.qq.com/s/FonPjJiI-H9unH8C50kI0w",
  },
  {
    title: "收藏！2025最新政策盘点！孤独症家庭必须了解的福利申领和课程学习！",
    date: "2025-04-24",
    href: "https://mp.weixin.qq.com/s/RrFhjMmWxDv_9a3ufvRODA",
  },
  {
    title: "孤独症患者医保政策及支持措施",
    date: "2025-03-26",
    href: "https://mp.weixin.qq.com/s/ksYP0IU4H4AqZvZ1i_5kJQ",
  },
  {
    title: "孤独症儿童康复政策盘点与展望：支持措施与未来发展方向",
    date: "2025-01-03",
    href: "https://mp.weixin.qq.com/s/7IN0zWd8DZlazLnh8gOR8A",
  },
  {
    title: "已开始执行！这些政策利好自闭症家庭",
    date: "2025-01-02",
    href: "https://mp.weixin.qq.com/s/jXijCU-iMnW59Rpngov5YQ",
  },
]);

onMounted(() => {
  resizeElement();
  document.getElementsByClassName("el-pagination__goto")[0].childNodes[0].nodeValue = "前往";
});

const resizeElement = function () {
  mainWinHeight.value = window.innerHeight - 114;
};
const handleSizeChange = (page) => {
  console.log(page);
}
</script>

<style lang="stylus">
@import "@/assets/css/single-common.styl"
.news-content {
  margin 0 auto;
  margin-top 90px;
  width 80%;

  .news-row {
    height 72px;
    font-size: 15px;
    border-top: 1px dashed #E3E3E3;
    border-bottom: 1px dashed #E3E3E3;
    line-height 72px;
    cursor pointer;

    img {
      width 20px;
      height 20px;
      position relative;
      top 25px;
    }
  }
}

.pagination {
  display flex;
  justify-content center;
  margin-top 50px;
}
</style>