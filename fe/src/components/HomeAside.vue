<template>
  <div id="main">
    <input id="search" :placeholder="lang.search" />
    <el-tree
      :data="data"
      :defaultExpandAll="true"
      @node-click="handleNodeClick"
    />
  </div>
</template>

<script setup>
import { useRouter } from "vue-router";
import { ref } from "vue";
import useGroupStore from "../stores/group";
import lang from "../i18n/i18n";
import { http } from "@/utils/axios";

const groupStore = useGroupStore();
const router = useRouter();
const data = ref([]);

http.get("/api/group").then((res) => {
  if (res.data) data.value = res.data;
});

const handleNodeClick = function (data) {
  if (data.tag != null) {
    groupStore.name = data.label;
    groupStore.tag = data.tag;
    router.push({
      name: "list",
    });
  }
};
</script>

<style scoped>
#main {
  width: 243px;
  background-color: #f1f1f1;
  height: 100%;
}

#search {
  background-color: #d6e7f7;
  width: 100%;
  height: 40px;
  padding-left: 10px;
  border: none;
  outline: none;
  font-size: 16px;
}

.el-tree {
  background-color: #f1f1f1;
}

.add_group {
  font-size: 14px;
  text-align: left;
  padding-left: 15px;
}
</style>