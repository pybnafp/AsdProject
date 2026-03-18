<template>
  <div class="app-background">
    <van-nav-bar
        :title="title"
        @click-left="showPicker = true"
        custom-class="navbar"
    >
      <template #left>
        <img style="width: 28px;height: 28px" src="@/assets/img/icon-arrows-left.png" alt="" @click="goBack"/>
      </template>
    </van-nav-bar>
    <div id="container">
      <van-list
          v-model:error="error"
          v-model:loading="loading"
          :finished="finished"
          error-text="请求失败，点击重新加载"
          finished-text="没有更多了"
          @load="onLoad"
      >
        <div class="news-list" v-for="(news,index) in newsList" :key="index">
          <a :href="news.href" target="_blank">
            <div class="title-row">
              <p>{{ news.title }}</p>
              <img src="@/assets/img/icon-arrows-right.png" alt=""/>
            </div>
            <p>{{ news.date }}</p>
          </a>
        </div>
      </van-list>
    </div>
  </div>
</template>
<script setup>
import {onMounted, ref} from 'vue';

const title = ref("政策指引");
const loading = ref(false)
const finished = ref(false)
const error = ref(false)
const showPicker = ref(false)
const offset = ref(0)
const limit = ref(10)
const goBack = () => {
  history.back();
}
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

});
const onLoad = () => {
  loading.value = false;
};
</script>

<style lang="stylus">
.app-background {
  height: 100%;

  #container {
    min-width: 0;
    flex: 1;
    padding: 46px 15px;
  }
}

.news-list {
  margin auto;
  border-top: 1px dashed #E3E3E3;
  border-bottom: 1px dashed #E3E3E3;
  padding 15px 0;

  .title-row {
    display flex;
    justify-content space-between;
    cursor pointer;
    margin-bottom 12px;

    p {
      margin-right 30px;
      font-size: 14px;
      letter-spacing: 2px;
      line-height 22px;
      color: #000000;
    }

    img {
      width 20px;
      height 20px;
      flex-shrink 1;
    }
  }

  p {
    font-size: 12px;
    line-height: 16px;
    color: #999797;
  }
}
</style>