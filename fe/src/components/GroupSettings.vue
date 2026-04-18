<template>
  <div class="settings-card">
    <div class="settings-header">
      <h3>{{ lang.email_folders }}</h3>
      <p class="settings-desc">{{ lang.email_folders_desc }}</p>
    </div>

    <div class="tree-container">
      <el-tree
        :expand-on-click-node="false"
        :data="data"
        :defaultExpandAll="true"
        class="custom-tree"
      >
        <template #default="{ node, data }">
          <div class="tree-node-content">
            <div class="node-label">
              <el-icon class="folder-icon"><Folder /></el-icon>
              <span v-if="data.id !== -1">{{ data.label }}</span>
              <el-input
                v-if="data.id === -1"
                v-model="data.label"
                size="small"
                :placeholder="lang.folder_name"
                class="node-input"
                @blur="onInputBlur(data)"
                @keyup.enter="onInputBlur(data)"
                ref="newInput"
                autofocus
              ></el-input>
            </div>
            
            <div class="node-actions" v-if="data.id !== 0">
              <el-button @click.stop="add(data)" size="small" type="primary" text bg class="action-btn">
                <el-icon><Plus /></el-icon>
              </el-button>
              <el-button @click.stop="del(node, data)" size="small" type="danger" text bg class="action-btn">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
          </div>
        </template>
      </el-tree>
    </div>

    <div class="form-actions">
      <el-button type="primary" @click="addRoot" class="add-root-btn" plain>
        <el-icon><Plus /></el-icon> {{ lang.add_group }}
      </el-button>
    </div>
  </div>
</template>

<script setup>
import { reactive } from "vue";
import lang from "../i18n/i18n";
import { http } from "@/utils/axios";
import { ElMessage } from "element-plus";
import { Folder, Delete, Plus } from "@element-plus/icons-vue";

const data = reactive([]);

http.get("/api/group").then((res) => {
  data.push(...res.data);
});

const del = function (node, dataObj) {
  if (dataObj.id !== -1) {
    http.post("/api/group/del", { id: dataObj.id }).then((res) => {
      if (res.errorNo !== 0) {
        ElMessage({ message: res.errorMsg, type: "error" });
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
          ElMessage({ message: res.errorMsg, type: "error" });
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

<style scoped>
.settings-card {
  padding: 0;
}

.settings-header {
  margin-bottom: 24px;
}

.settings-header h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--pm-text-primary);
  margin: 0 0 8px 0;
}

.settings-desc {
  font-size: 14px;
  color: var(--pm-text-secondary);
  margin: 0;
}

.tree-container {
  background: var(--pm-bg-secondary);
  border: 1px solid var(--pm-border-color);
  border-radius: var(--pm-radius-sm);
  padding: 16px;
  margin-bottom: 24px;
}

.custom-tree {
  background: transparent;
}

.tree-node-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  padding-right: 16px;
}

.node-label {
  display: flex;
  align-items: center;
  font-size: 14px;
  color: var(--pm-text-primary);
}

.folder-icon {
  margin-right: 8px;
  color: var(--pm-primary-color);
  font-size: 16px;
}

.node-input {
  width: 200px;
  margin-left: 8px;
}

.node-actions {
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.2s;
}

.tree-node-content:hover .node-actions {
  opacity: 1;
}

.action-btn {
  padding: 4px 8px;
  height: 24px;
}

.form-actions {
  display: flex;
  justify-content: flex-start;
}

.add-root-btn {
  border-radius: var(--pm-radius-sm);
}

/* Override element-plus tree hover styles */
:deep(.el-tree-node__content) {
  height: 40px;
  border-radius: var(--pm-radius-sm);
  margin-bottom: 4px;
}

:deep(.el-tree-node__content:hover) {
  background-color: var(--pm-bg-hover);
}
</style>
