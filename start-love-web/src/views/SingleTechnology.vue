<template>
  <div class="single-page">
    <el-container>
      <el-header>
        <header-bar></header-bar>
      </el-header>
      <el-main>
        <div id="container" :style="{ height: mainWinHeight + 'px' }">
          <div class="content-card" :style="{ height: cardHeight + 'px', background: '#FCF7FF'}">
            <div class="card-bg" style="display: flex;flex-direction: column; justify-content: space-between">
              <div style="font-size: 24px; padding-top: 50px">
                为孤独症群体、家长、医生等用户提供早筛分析、专业问答、诊疗建议、研究进展等权威且及时的内容服务。
              </div>
              <img style="max-height:80%;object-fit:cover; margin: 0 auto 50px;" src="/images/illustration/bg1.png"
                   alt="">
            </div>
          </div>
          <div class="content-card" :style="{ height: cardHeight + 'px' }">
            <div class="card-layout card-bg">
              <img src="/images/illustration/bg2.png" alt="">
              <div class="intro">
                <p class="title">多种模态的AI早筛技术</p>
                <p class="text">ASD具有高度遗传性，估计遗传因素贡献率达80%以上可在出生<span
                    style="color: #5D19D2">甚至产前检测</span>。可辅助早期识别高风险婴儿特别是在有家族史的个体中。与其他早筛手段结合可提升预测准确性。
                </p>
              </div>
            </div>
          </div>
          <div class="content-card" :style="{ height: cardHeight + 'px' }">
            <div class="card-layout card-bg">
              <div class="intro">
                <p class="title">脑影像技术</p>
                <p class="text">可以在婴儿甚至新生儿阶段，识别出脑结构或功能的异常，为早筛提供了可能。</p>
              </div>
              <img src="/images/illustration/bg3.png" alt="">
            </div>
          </div>
          <div class="content-card" :style="{ height: cardHeight + 'px' }">
            <div class="card-layout card-bg">
              <img src="/images/illustration/bg4.png" alt="">
              <div class="intro">
                <p class="title">行为视频训练</p>
                <p class="text">
                  基于孤独症人群的面容和行为视频训练出的面容和行为模型。通过面部表情、行为特征等辅助判断ASD风险。</p>
              </div>
            </div>
          </div>
          <div class="content-card" :style="{ height: cardHeight + 'px' }">
            <div class="card-layout card-bg">
              <div class="intro">
                <p class="title">眼动追踪</p>
                <p class="text">
                  眼动追踪是一种非侵入式、客观且高效的行为测量方法，尤其适用于婴幼儿和语言沟通能力弱的群体。优势在于客观数据、抗主观偏差，可量化评估社交注意模式。</p>
              </div>
              <img src="/images/illustration/bg5.png" alt="">
            </div>
          </div>
          <div class="content-card" :style="{ height: cardHeight + 'px' }">
            <div class="card-layout card-bg">
              <img src="/images/illustration/bg6.png" alt="">
              <div class="intro">
                <p class="title">孤独症问卷量表</p>
                <p class="text">AI技术分析孤独症问卷量表，可以帮助提高诊断效率、识别关键特征、进行个性化干预建议</p>
              </div>
            </div>
          </div>
          <div class="content-card" :style="{ height: cardHeight + 'px' }">
            <div class="card-layout card-bg">
              <div class="intro">
                <p class="title">大语言模型</p>
                <p class="text">传统大语言模型（LLM）虽然强大，但存在<span
                    style="color: #5D19D2">知识更新滞后</span>和<span style="color: #5D19D2">缺乏准确性保障的问题</span>。为解决这一点，结合<span
                    style="color: #5D19D2">向量数据库（向量检索）实现专业知识增强问答</span>。</p>
              </div>
              <img src="/images/illustration/bg7.png" alt="">
            </div>
          </div>
          <div class="content-card" :style="{ height: cardHeight + 'px' }">
            <div class="card-layout card-bg">
              <img src="/images/illustration/bg8.png" alt="">
              <div class="intro">
                <p class="title">知识库</p>
                <p class="text">知识图谱可构建结构化知识库，提升专业性与准确率。支持复杂、多跳问题回答</p>
              </div>
            </div>
          </div>
          <div class="content-card" :style="{ height: cardHeight + 'px' }">
            <div class="card-layout card-bg">
              <div class="intro">
                <p class="title">软硬件一体的解决方案</p>
                <p class="text">
                  高精度眼动设备采集眼动、面容、行为信息。结合先进AI算法评估认知功能并给出ASD风险提示。性价比高，耗时短，自动采集，可用于广泛筛查阶段。结合后续的脑影像、基因早筛等技术构建多级筛查流程。</p>
              </div>
              <img src="/images/illustration/bg9.png" alt="">
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
const cardHeight = ref(0);

onMounted(() => {
  resizeElement();
});

const resizeElement = function () {
  mainWinHeight.value = window.innerHeight - 114;
  cardHeight.value = window.innerHeight - 150;
};
</script>
<style lang="stylus">
@import "@/assets/css/single-common.styl"
.content-card {
  border-radius: 24px;
  box-shadow: 0px 4px 40px 0px #F0EDF6;
  padding 0 50px;
  text-align center;
  margin 30px auto;
  width 80%;

}

.card-layout {
  display flex;
  justify-content center;
  align-items center;
  gap 10px;
  img {
    border-radius: 47px;
    width 50%;
  }

  .intro {
    padding 50px;
    text-align justify;

    .title {
      font-size 30px;
      font-weight: bold;
    }

    .text {
      margin-top 40px;
      font-size: 22px;
      font-weight: 500;
      line-height: 40px;
    }
  }
}
</style>