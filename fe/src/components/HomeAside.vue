<template>
  <div id="main">
    <input id="search" :placeholder="lang.search">
    <el-tree :data="data" :props="defaultProps" :defaultExpandAll="true" @node-click="handleNodeClick" :class="node" />
  </div>
</template>


<script setup>
import { useRouter } from 'vue-router'
import $http from "../http/http";
import { reactive, ref } from 'vue'
import useGroupStore from '../stores/group'
import lang from '../i18n/i18n';

const groupStore = useGroupStore()
const router = useRouter()


const data = ref([])

$http.get('/api/group').then(res => {
  data.value = res.data
})




const handleNodeClick = function (data) {
  if (data.tag != null) {
    groupStore.name = data.label
    groupStore.tag = data.tag
    router.push({
      name: "list",
    })
  }
}



</script>


<style scoped>
#main {
  width: 243px;
  background-color: #F1F1F1;
  height: 100%;
}

#search {
  background-color: #D6E7F7;
  width: 100%;
  height: 40px;
  padding-left: 10px;
  border: none;
  outline: none;
  font-size: 16px;
}

.el-tree {
  background-color: #F1F1F1;
}
</style>