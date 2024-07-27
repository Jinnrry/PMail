<template>
    <div id="header_main">
        <div id="logo">
            <span style="padding-left: 20px;">PMail</span>
        </div>
        <div id="settings" @click="settings" v-if="$isLogin">
            <el-icon style="font-size: 25px;">
                <Setting style="color:#FFFFFF" />
            </el-icon>
        </div>
        <el-drawer v-model="openSettings" size="80%" :title="lang.settings">
            <el-tabs tab-position="left">
                <el-tab-pane :label="lang.security">
                    <SecuritySettings />
                </el-tab-pane>

                <el-tab-pane :label="lang.group_settings">
                    <GroupSettings />
                </el-tab-pane>

                <el-tab-pane :label="lang.rule_setting">
                    <RuleSettings />
                </el-tab-pane>

                <el-tab-pane v-if="$userInfos.is_admin" :label="lang.user_management">
                    <UserManagement />
                </el-tab-pane>

                <el-tab-pane :label="lang.plugin_settings">
                    <PluginSettings />
                </el-tab-pane>

            </el-tabs>
        </el-drawer>

    </div>
</template>

<script setup>
import { Setting } from '@element-plus/icons-vue';
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import SecuritySettings from '@/components/SecuritySettings.vue'
import lang from '../i18n/i18n';
import GroupSettings from './GroupSettings.vue';
import RuleSettings from './RuleSettings.vue';
import UserManagement from './UserManagement.vue';
import { getCurrentInstance } from 'vue'
import PluginSettings from './PluginSettings.vue';
const app = getCurrentInstance()
const $http = app.appContext.config.globalProperties.$http
const $isLogin = app.appContext.config.globalProperties.$isLogin
const $userInfos = app.appContext.config.globalProperties.$userInfos

const openSettings = ref(false)
const settings = function () {
    if (Object.keys($userInfos.value).length == 0) {
        $http.post("/api/user/info", {}).then(res => {
            if (res.errorNo == 0) {
                $userInfos.value = res.data
                openSettings.value = true;

            } else {
                ElMessage({
                    type: 'error',
                    message: res.errorMsg,
                })
            }
        })
    }else{
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