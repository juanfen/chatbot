<script setup lang="ts">
import GlobalContent from '../modules/global-content/index.vue';

defineOptions({
  name: 'BlankLayout'
});
</script>

<template>
  <!-- 与 AdminLayout 不同，此处无外层 flex 主区域；须自带高度链，否则登录/异常页内 size-full、h-full 会塌缩 -->
  <div class="box-border h-full min-h-screen flex flex-col">
    <GlobalContent :show-padding="false" />
  </div>
</template>

<style scoped></style>
