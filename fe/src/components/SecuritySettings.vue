<template>
    <el-form :model="ruleForm" :rules="rules" status-icon>
        <el-form-item :label="lang.modify_pwd" prop="new_pwd">
            <el-input type="password" v-model="ruleForm.new_pwd" />
        </el-form-item>

        <el-form-item :label="lang.enter_again" prop="new_pwd2">
            <el-input type="password" v-model="ruleForm.new_pwd2" />
        </el-form-item>

        <el-form-item>
            <el-button type="primary" @click="submit">
                {{ lang.submit }}
            </el-button>
        </el-form-item>
    </el-form>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { ElNotification } from 'element-plus'
import $http from "../http/http";
import lang from '../i18n/i18n';
const ruleForm = reactive({
    new_pwd: "",
    new_pwd2: ""
})

const rules = reactive({
    new_pwd: [{ required: true, message: lang.err_required_pwd, trigger: 'blur' },],
    new_pwd2: [{ required: true, message: lang.err_required_pwd, trigger: 'blur' },],

})

const submit = function () {
    if (ruleForm.new_pwd != ruleForm.new_pwd2) {
        ElNotification({
            title: 'Error',
            message: lang.err_pwd_diff,
            type: 'error',
        })
        return
    }
    $http.post("/api/settings/modify_password", { password: ruleForm.new_pwd }).then(res => {
        ElNotification({
            title: res.errorNo == 0 ? lang.succ : lang.fail,
            message: res.data,
            type: res.errorNo == 0 ? 'success' : 'error',
        })
    })


}
</script>