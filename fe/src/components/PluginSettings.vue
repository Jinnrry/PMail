<template>
  <div class="settings-card">
    <div class="settings-header">
      <h3>{{ lang.extensions }}</h3>
      <p class="settings-desc">{{ lang.extensions_desc }}</p>
    </div>

    <div class="plugin-container">
      <el-tabs class="custom-tabs">
        <el-tab-pane v-for="(src, name) in pluginList" :key="src" :label="name">
          <div class="iframe-wrapper">
            <iframe :src="src"></iframe>
          </div>
        </el-tab-pane>
        <el-tab-pane v-if="Object.keys(pluginList).length === 0" label="No Plugins">
          <div class="empty-state">
            <el-empty description="No plugins are currently installed" :image-size="120" />
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script setup>
import {reactive} from 'vue'
import {http} from "@/utils/axios";
import lang from '../i18n/i18n';

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
.settings-card {
  height: 100%;
  display: flex;
  flex-direction: column;
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

.plugin-container {
  flex-grow: 1;
  border: 1px solid var(--pm-border-color);
  border-radius: var(--pm-radius-sm);
  background: var(--pm-bg-secondary);
  overflow: hidden;
}

.custom-tabs :deep(.el-tabs__header) {
  margin-bottom: 0;
  background: var(--pm-bg-primary);
  padding: 0 16px;
  border-bottom: 1px solid var(--pm-border-color);
}

.custom-tabs :deep(.el-tabs__content) {
  height: 100%;
}

.iframe-wrapper {
  height: calc(100vh - 250px);
  min-height: 400px;
}

iframe {
  width: 100%;
  height: 100%;
  border: 0;
}

.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 300px;
}
</style>