<template>
    <div id="main">
        <div id="form">
            <el-form :model="form" label-width="120px">
                <el-form-item :label="lang.account">
                    <el-input v-model="form.account" placeholder="User Name" />
                </el-form-item>
                <el-form-item :label="lang.password">
                    <el-input v-model="form.password" placeholder="Password" type="password" />
                </el-form-item>
                <el-form-item>
                    <el-button type="primary" @click="onSubmit">{{ lang.login }}</el-button>
                </el-form-item>
            </el-form>

        </div>
    </div>
</template>

<script setup>

import { reactive } from 'vue'
import $http from "../http/http";
import { ElMessage } from 'element-plus'
import router from "@/router";  //根路由对象
import lang from '../i18n/i18n';


const form = reactive({
    account: '',
    password: '',
})

const onSubmit = () => {
    $http.post("/api/login", form).then(res => {
        if (res.errorNo != 0) {
            ElMessage.error(res.errorMsg)
        } else {
            router.replace({
                path: '/',
                query: {
                    redirect: router.currentRoute.fullPath
                }
            })
        }
    })

}
</script>


<style scoped>
#main {
    width: 100%;
    height: 100%;
    background-color: #f1f1f1;
    display: flex;
    justify-content: center;
    /* 水平居中 */
    align-items: center;
    /* 垂直居中 */
}
</style>