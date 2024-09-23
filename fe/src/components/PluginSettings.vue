<template>
  <div id="main">
    <el-tabs>
      <el-tab-pane v-for="(src, name) in pluginList" :key="src" :label="name">
        <iframe :src="src"></iframe>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import {reactive} from 'vue'
import {http} from "@/utils/axios";

const pluginList = reactive({})

http.get('/api/plugin/list').then(res => {
  if (res.data != null && res.data.length > 0) {
    for (let i = 0; i < res.data.length; i++) {
      let name = res.data[i];
      pluginList[name] = "/api/plugin/settings/" + name + "/index.html";
    }
  }
})

</script>

<style scoped>

iframe {
  width: 100%;
  border: 0;
}

</style>