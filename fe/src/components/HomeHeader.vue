<template>
  <div id="header_main">
    <div id="logo">
      <router-link to="/">
        <el-text :line-clamp="1" size="large"><h1>Pmail</h1></el-text>
      </router-link>
    </div>
    <div id="settings" @click="settings" v-if="isLogin">
      <el-icon style="font-size: 25px;">
        <TbSettings style="color:#FFFFFF"/>
      </el-icon>
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
import {TbSettings} from "vue-icons-plus/tb";
import {ref} from 'vue'
import {ElMessage} from 'element-plus'
import SecuritySettings from '@/components/SecuritySettings.vue'
import lang from '../i18n/i18n';
import GroupSettings from './GroupSettings.vue';
import RuleSettings from './RuleSettings.vue';
import UserManagement from './UserManagement.vue';
import PluginSettings from './PluginSettings.vue';
import {http} from "@/utils/axios";
import {useGlobalStatusStore} from "@/stores/useGlobalStatusStore";

const globalStatus = useGlobalStatusStore();
const isLogin = globalStatus.isLogin;
const userInfos = globalStatus.userInfos;


const openSettings = ref(false)
const settings = function () {
  if (Object.keys(userInfos).length === 0) {
    http.post("/api/user/info", {}).then(res => {
      if (res.errorNo === 0) {
        userInfos.value = res.data
        openSettings.value = true;
      } else {
        ElMessage({
          type: 'error',
          message: res.errorMsg,
        })
      }
    })
  } else {
    openSettings.value = true;
  }


}

</script>


<style scoped>

#header_main {
  height: 50px;
  background-color: #000;
  display: flex;
  padding: 0;
}

#logo {
  height: 3rem;
  line-height: 3rem;
  font-size: 2.3rem;
  flex-grow: 1;
  width: 200px;
  color: #FFF;
  text-align: left;
}

#logo h1 {
  padding-left: 20px;
  color: white;
}

#search {
  height: 3rem;
  width: 100%;
}

#settings {
  display: flex;
  justify-content: center;
  align-items: center;
  padding-right: 20px;
}
</style>