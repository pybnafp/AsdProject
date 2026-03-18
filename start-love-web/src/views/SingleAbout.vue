<template>
  <div class="single-page">
    <el-container>
      <el-header>
        <header-bar></header-bar>
      </el-header>
      <el-main>
        <div id="container" :style="{ height: mainWinHeight + 'px' }">
          <div class="text-box">
            <img src="/images/illustration/bg16.png" alt="">
            <div class="text-paragraph">
              <h2>我是谁？</h2>
              <span class="text-content">星启诶艾StellarCare AI（SCAI）是首个融合多模态早筛数据、基于大语言模型和专业知识库的孤独症AI智能体，致力于实现孤独症“早筛—诊断—干预—保障”全链条闭环。SCAI努力推动社会协同，构建孤独症全生涯支持平台。</span>
            </div>
            <div class="text-paragraph">
              <h2>从哪来？</h2>
              <ul class="text-content" style="padding-left: 20px;">
                <li>2024年11月,福建省与你孤独症儿童服务中心与中科院杭州医学所联合发起“AI+孤独症早筛课题”。</li>
                <li>2025年3月27日，课题组在泉州开展多模态数据试点采集。</li>
                <li>2025年6月1日，首个孤独症 Al智能体 StellarCare Al 正式发布。</li>
              </ul>
            </div>
            <div class="text-paragraph">
              <h2>做什么？</h2>
              <ul class="text-content" style="padding-left: 20px;">
                <li>-更智慧的早筛方案：融合量表、眼动、基因等多模态数据，准确率超85%，诊断提前至18月龄；推动分级筛查，提升效率；协作进化，持续优化效果。</li>
                <li>-更开放的知识生态：汇聚全球前沿研究与权威指南，构建中国孤独症多模态数据库，持续更新进化的孤独症知识体系。</li>
                <li>-更温暖的协同平台：链接家庭、医疗与社会，为家长减压、医生赋能、社会共治，打造全方位支持体系。</li>
              </ul>
            </div>
            <div class="text-paragraph">
              <h2>联系我们：stellarcareai@126.com</h2>
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

onMounted(() => {
  resizeElement();
});

const resizeElement = function () {
  mainWinHeight.value = window.innerHeight - 114;
};
</script>

<style lang="stylus">
@import "@/assets/css/single-common.styl"

.text-box {
  width 80%;
  margin 50px auto;
  .text-paragraph {
    text-align justify;
    margin 40px auto;
    .text-content {
      font-size: 18px;
      line-height: 40px;
      color: #9E9E9E;
      list-style-type: disc;
    }
  }
}
</style>