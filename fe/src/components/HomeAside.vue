<template>
  <div class="aside-container">
    <div class="search-bar">
      <el-input :placeholder="lang.search" />
    </div>
    <el-scrollbar>
      <el-tree
          :data="data"
          :default-expand-all="true"
          @node-click="handleNodeClick"
          class="custom-tree"
      />
    </el-scrollbar>
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
.aside-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: #fafafa;
  border-right: 1px solid #e0e0e0;
}

.search-bar {
  padding: 10px;
  border-bottom: 1px solid #e0e0e0;
}

.el-scrollbar {
  flex-grow: 1;
}

.custom-tree {
  background-color: transparent;
  padding: 10px;
}

.custom-tree >>> .el-tree-node__content {
  height: 36px;
  line-height: 36px;
  border-radius: 4px;
}

.custom-tree >>> .el-tree-node__content:hover {
  background-color: #ecf5ff;
}

.custom-tree >>> .el-tree-node:focus > .el-tree-node__content {
  background-color: #d9ecff;
}

.custom-tree >>> .el-tree-node.is-current > .el-tree-node__content {
  background-color: #d9ecff;
  color: #409eff;
  font-weight: bold;
}
</style>
