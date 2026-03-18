<template>
  <el-container class="chat-file-list">
    <div v-for="file in fileList" :key="file.file_id">
      <div class="item">
        <img :src="GetFileIcon(file.file_type)" alt="">
        <div class="body">
          <div class="title">
            {{ substr(file.file_name, 30) }}
          </div>
          <div class="info">
            <span>{{ GetFileType(file.file_type) }}·</span>
            <span>{{ FormatFileSize(file.file_size) }}</span>
          </div>
        </div>
        <div class="action">
          <el-icon @click="removeFile(file)">
            <CircleCloseFilled/>
          </el-icon>
        </div>
      </div>
    </div>
  </el-container>
</template>

<script setup>
import {ref} from "vue";
import {CircleCloseFilled} from "@element-plus/icons-vue";
import {removeArrayItem, substr} from "@/utils/libs";
import {FormatFileSize, GetFileType, GetFileIcon} from "@/store/system";

const props = defineProps({
  files: {
    type: Array,
    default: [],
  },
});
const emits = defineEmits(["removeFile"]);
const fileList = ref(props.files);

const removeFile = (file) => {
  fileList.value = removeArrayItem(fileList.value, file, (v1, v2) => v1.url === v2.url);
  emits("removeFile", file);
};
</script>

<style scoped lang="stylus">

.chat-file-list {
  display flex
  flex-flow row

  .item {
    position relative
    display flex
    flex-flow row
    border-radius 12px
    background-color: #F8F5FE;
    padding 12px 12px;
    margin-right 10px;


    img {
      width 44px
      height 44px
    }

    .body {
      margin-left 8px
      text-align left

      .title {
        font-size 16px
      }

      .info {
        font-size: 12px;
        color #6E6E6E;
        margin-top 3px;
      }
    }
  }

  .action {
    position absolute
    top -8px
    right -8px
    color #da0d54
    cursor pointer
    font-size 20px

    .el-icon {
      background-color #fff
      border-radius 50%
    }
  }
}
</style>
