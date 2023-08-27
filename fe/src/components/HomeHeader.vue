<template>
    <div id="header_main">
        <div id="logo">
            <span style="padding-left: 20px;">PMail</span>
        </div>
        <div id="settings" @click="settings">
            <el-icon style="font-size: 25px;">
                <Setting style="color:#FFFFFF" />
            </el-icon>
        </div>
        <el-drawer v-model="openSettings" size="80%" :title="lang.settings">
            <el-tabs tab-position="left" >
                <el-tab-pane :label="lang.security">
                    <SecuritySettings/>
                </el-tab-pane>

                <el-tab-pane :label="lang.group_settings">
                    <GroupSettings/>
                </el-tab-pane>
            </el-tabs>
        </el-drawer>

    </div>
</template>

<script setup>
import { Setting } from '@element-plus/icons-vue';
import { ref } from 'vue'
import { ElMessageBox } from 'element-plus'
import SecuritySettings from '@/components/SecuritySettings.vue'
import lang from '../i18n/i18n';
import GroupSettings from './GroupSettings.vue';


const openSettings = ref(false)
const settings = function () {
    openSettings.value = true;
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