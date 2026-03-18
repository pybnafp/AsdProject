<template>
  <div>
    <div class="menu-container">
      <div class="menu-list">
        <el-link v-for="(item, index) in linkList" :href="item.href" underline="never" :type="item.href === curPath ? 'primary' : ''" :key="index">{{item.title}}</el-link>
      </div>
    </div>
  </div>
</template>
<script setup>
import { useRouter } from "vue-router";
import { ref,} from "vue";

const router = useRouter();
const curPath = ref("");

const linkList = ref([
  {
    title: "早筛技术",
    href: "/single-page/early-screening-technology",
  },
  {
    title: "政策指引",
    href: "/single-page/policy-guide",
  },
  {
    title: "星启协爱",
    href: "/single-page/star-love",
  },
  {
    title: "公益机构",
    href: "/single-page/public-institution",
  },
  {
    title: "开放平台",
    href: "/single-page/open-platform",
  },
  {
    title: "关于我们",
    href: "/single-page/about",
  },
]);

curPath.value = router.currentRoute.value.path;

// 监听路由变化;
router.beforeEach((to, from, next) => {
  curPath.value = to.path;
  next();
});
</script>
<style scoped lang="stylus">

.menu-container {
  width: 100%;

  .menu-list {
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-radius: 24px;
    background: #FFFFFF;
    box-shadow: 0 17px 40px 0 #F0EDF6;
    max-width: 1004px;
    height: 64px;
    font-size: 16px;
    font-weight: normal;
    line-height: 18px;
    padding: 0 132px;
    margin auto;
  }
}

</style>