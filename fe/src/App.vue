<script setup>
import {RouterView, useRoute} from 'vue-router'
import HomeHeader from '@/components/HomeHeader.vue'
import HomeAside from '@/components/HomeAside.vue';
import {ref, watch} from 'vue'

const route = useRoute()
const pageName = ref(route.name)

watch(
    () => route.fullPath,
    () => {
      pageName.value = route.name
    }
)

</script>

<template>
  <el-container class="main-container">
    <el-header class="header">
      <HomeHeader/>
    </el-header>
    <el-container>
      <el-aside width="240px" v-if="pageName !== 'login' && pageName !== 'setup'">
        <HomeAside/>
      </el-aside>
      <el-main>
        <RouterView/>
      </el-main>
    </el-container>
  </el-container>
</template>


<style>
html, body, #app, .main-container {
  height: 100%;
  margin: 0;
  padding: 0;
}

.header {
  padding: 0 !important;
  height: auto !important;
  border-bottom: 1px solid #dcdfe6;
}

.el-main {
  padding: 0 !important;
}
</style>
