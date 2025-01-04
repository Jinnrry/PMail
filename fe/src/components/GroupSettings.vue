<template>
  <div id="main">
    <el-tree
      :expand-on-click-node="false"
      :data="data"
      :defaultExpandAll="true"
    >
      <template #default="{ node, data }">
        <div>
          <span v-if="data.id !== -1"> {{ data.label }}</span>
          <el-input
            v-if="data.id === -1"
            v-model="data.label"
            @blur="onInputBlur(data)"
          ></el-input>
          <el-button
            v-if="data.id !== 0"
            @click="del(node, data)"
            size="small"
            type="danger"
            circle
          >
            -
          </el-button>
          <el-button
            v-if="data.id !== 0"
            @click="add(data)"
            size="small"
            type="primary"
            circle
          >
            +
          </el-button>
        </div>
      </template>
    </el-tree>

    <el-button @click="addRoot">{{ lang.add_group }}</el-button>
  </div>
</template>

<script setup>
import { reactive } from "vue";
import lang from "../i18n/i18n";
import { http } from "@/utils/axios";
import { ElMessage } from "element-plus";

const data = reactive([]);

http.get("/api/group").then((res) => {
  data.push(...res.data);
});

const del = function (node, data) {
  if (data.id !== -1) {
    http.post("/api/group/del", { id: data.id }).then((res) => {
      if (res.errorNo !== 0) {
        ElMessage({
          message: res.errorMsg,
          type: "error",
        });
      } else {
        const pc = node.parent.childNodes;
        for (let i = 0; i < pc.length; i++) {
          if (pc[i].id === node.id) {
            pc.splice(i, 1);
            return;
          }
        }
      }
    });
  } else {
    const pc = node.parent.childNodes;
    for (let i = 0; i < pc.length; i++) {
      if (pc[i].id === node.id) {
        pc.splice(i, 1);
        return;
      }
    }
  }
};

const add = function (item) {
  if (item.children == null) {
    item.children = [];
  }
  item.children.push({
    children: [],
    label: "",
    id: -1,
    parent_id: item.id,
  });
};

const addRoot = function () {
  data.push({
    children: [],
    label: "",
    id: -1,
    parent_id: 0,
  });
};

const onInputBlur = function (item) {
  if (item.label !== "") {
    http.post("/api/group/add", { name: item.label, parent_id: item.parent_id })
      .then((res) => {
        if (res.errorNo !== 0) {
          ElMessage({
            message: res.errorMsg,
            type: "error",
          });
        } else {
          http.get("/api/group").then((res) => {
            data.splice(0, data.length);
            data.push(...res.data);
          });
        }
      });
  }
};
</script>
