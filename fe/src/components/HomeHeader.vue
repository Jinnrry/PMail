<template>
  <div class="header-container">
    <div class="logo-section">
      <router-link to="/" class="logo-link">
        <img src="@/assets/logo.svg" alt="PMail Logo" class="logo-img"/>
        <span class="logo-text">PMail</span>
      </router-link>
    </div>
    <div class="actions-section" v-if="isLogin">
      <el-button :icon="Setting" circle @click="settings"/>
    </div>
    <el-drawer v-model="openSettings" size="80%" :title="lang.settings">
      <el-tabs tab-position="left">
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
  </div>
</template>

<script setup>
import {Setting} from "@element-plus/icons-vue";
import {ref} from 'vue'
import SecuritySettings from '@/components/SecuritySettings.vue'
import lang from '../i18n/i18n';
import GroupSettings from './GroupSettings.vue';
import RuleSettings from './RuleSettings.vue';
import UserManagement from './UserManagement.vue';
import PluginSettings from './PluginSettings.vue';
import {useGlobalStatusStore} from "@/stores/useGlobalStatusStore";

const globalStatus = useGlobalStatusStore();
const isLogin = globalStatus.isLogin;
const userInfos = globalStatus.userInfos;


const openSettings = ref(false)
const settings = function () {
  if (Object.keys(userInfos).length === 0) {
    globalStatus.init(()=>{
      Object.assign(userInfos,globalStatus.userInfos)
      openSettings.value = true;
    })
  } else {
    openSettings.value = true;
  }


}

</script>


<style scoped>
.header-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 60px;
  padding: 0 20px;
  background-color: #ffffff;
}

.logo-section {
  display: flex;
  align-items: center;
}

.logo-link {
  display: flex;
  align-items: center;
  text-decoration: none;
}

.logo-img {
  height: 32px;
  width: 32px;
  margin-right: 10px;
}

.logo-text {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}

.actions-section {
  display: flex;
  align-items: center;
}
</style>
