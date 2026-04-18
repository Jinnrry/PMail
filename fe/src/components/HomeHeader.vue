<template>
  <div class="header-container">
    <div class="header-left">
      <div class="mobile-menu-btn" @click="globalStatus.mobileDrawerVisible = true">
        <el-icon><EpMenu /></el-icon>
      </div>
      <div class="mobile-logo">
        <router-link to="/">
          <h1>PMail</h1>
        </router-link>
      </div>
    </div>
    
    <div class="header-right">
    </div>

  </div>

  <!-- Settings Drawer -->
  <el-drawer v-model="globalStatus.settingsDrawerVisible" :size="isMobile ? '100%' : '600px'" :title="lang.settings" class="settings-drawer" :with-header="true" direction="rtl">
    <el-tabs :tab-position="isMobile ? 'top' : 'left'" class="settings-tabs">
      <el-tab-pane :label="lang.security">
        <SecuritySettings/>
      </el-tab-pane>
      <el-tab-pane :label="lang.group_settings">
        <GroupSettings/>
      </el-tab-pane>
      <el-tab-pane :label="lang.rule_setting">
        <RuleSettings/>
      </el-tab-pane>
      <el-tab-pane v-if="userInfos.is_admin" :label="lang.user_management">
        <UserManagement/>
      </el-tab-pane>
      <el-tab-pane :label="lang.plugin_settings">
        <PluginSettings/>
      </el-tab-pane>
    </el-tabs>
  </el-drawer>
</template>

<script setup>
import {EpMenu} from "vue-icons-plus/ep";
import {ref, onMounted, onUnmounted} from 'vue'
import SecuritySettings from '@/components/SecuritySettings.vue'
import lang from '../i18n/i18n';
import GroupSettings from './GroupSettings.vue';
import RuleSettings from './RuleSettings.vue';
import UserManagement from './UserManagement.vue';
import PluginSettings from './PluginSettings.vue';
import {useGlobalStatusStore} from "@/stores/useGlobalStatusStore";

const globalStatus = useGlobalStatusStore();
const userInfos = globalStatus.userInfos;

const isMobile = ref(window.innerWidth <= 768);

const handleResize = () => {
  isMobile.value = window.innerWidth <= 768;
};

onMounted(() => {
  window.addEventListener('resize', handleResize);
});

onUnmounted(() => {
  window.removeEventListener('resize', handleResize);
});
</script>

<style scoped>
.header-container {
  display: none;
}

.header-left, .header-right {
  display: flex;
  align-items: center;
}

.mobile-menu-btn {
  display: none;
  width: 36px;
  height: 36px;
  justify-content: center;
  align-items: center;
  font-size: 18px;
  color: var(--pm-text-secondary);
  cursor: pointer;
  border-radius: 12px;
  border: 1px solid var(--pm-border-color);
  background: var(--pm-surface-glass);
  transition: all 0.2s;
}

.mobile-menu-btn:hover {
  background-color: var(--pm-surface-solid);
  color: var(--pm-text-primary);
}

.mobile-logo {
  display: none;
  margin-left: 12px;
}

.mobile-logo h1 {
  font-size: 20px;
  font-weight: 600;
  color: var(--pm-primary-color);
  margin: 0;
  letter-spacing: -0.02em;
}

.mobile-logo a {
  text-decoration: none;
}

.settings-btn {
  font-size: 22px;
  color: var(--pm-text-secondary);
  cursor: pointer;
  padding: 8px;
  border-radius: var(--pm-radius-md);
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
}

.settings-btn:hover {
  color: var(--pm-primary-color);
  background-color: var(--el-color-primary-light-9);
}

.settings-tabs {
  height: 100%;
}

@media (max-width: 768px) {
  .header-container {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin: 10px 10px 0;
    padding: 12px 14px;
    border-radius: var(--pm-radius-lg);
    border: 1px solid var(--pm-border-color);
    background: var(--pm-surface-glass);
    backdrop-filter: blur(12px);
  }
  .mobile-menu-btn {
    display: flex;
  }
  .mobile-logo {
    display: block;
  }
}
</style>