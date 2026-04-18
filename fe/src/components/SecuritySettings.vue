<template>
  <div class="settings-card">
    <div class="settings-header">
      <h3>{{ lang.modify_pwd }}</h3>
      <p class="settings-desc">{{ lang.modify_pwd_desc }}</p>
    </div>
    
    <div class="settings-body">
      <el-form :model="ruleForm" :rules="rules" status-icon label-position="top" class="settings-form">
        <el-form-item :label="lang.modify_pwd" prop="new_pwd">
          <el-input type="password" v-model="ruleForm.new_pwd" :placeholder="lang.enter_new_pwd" show-password class="premium-input"/>
        </el-form-item>

        <el-form-item :label="lang.enter_again" prop="new_pwd2">
          <el-input type="password" v-model="ruleForm.new_pwd2" :placeholder="lang.confirm_new_pwd" show-password class="premium-input"/>
        </el-form-item>

        <div class="form-actions">
          <el-button type="primary" @click="submit" class="submit-btn">
            {{ lang.submit }}
          </el-button>
        </div>
      </el-form>
    </div>

    <el-divider class="settings-divider"/>

    <div class="settings-header">
      <h3 class="danger-text">{{ lang.logout }}</h3>
      <p class="settings-desc">{{ lang.logout_desc }}</p>
    </div>
    <div class="settings-body">
      <el-button type="danger" plain @click="logout" class="logout-btn">
        {{ lang.logout }}
      </el-button>
    </div>
  </div>
</template>

<script setup>
import {reactive} from 'vue'
import {ElNotification} from 'element-plus'
import lang from '../i18n/i18n';
import {http} from "@/utils/axios";

const ruleForm = reactive({
  new_pwd: "",
  new_pwd2: ""
})

const rules = reactive({
  new_pwd: [{required: true, message: lang.err_required_pwd, trigger: 'blur'},],
  new_pwd2: [{required: true, message: lang.err_required_pwd, trigger: 'blur'},],
})

const logout = function () {
  http.post("/api/logout", {}).then(() => {
    location.reload();
  })
}

const submit = function () {
  if (ruleForm.new_pwd === "") return;
  if (ruleForm.new_pwd !== ruleForm.new_pwd2) {
    ElNotification({
      title: 'Error',
      message: lang.err_pwd_diff,
      type: 'error',
    })
    return
  }
  http.post("/api/settings/modify_password", {password: ruleForm.new_pwd}).then(res => {
    ElNotification({
      title: res.errorNo === 0 ? lang.succ : lang.fail,
      message: res.data,
      type: res.errorNo === 0 ? 'success' : 'error',
    })
  })
}
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

.danger-text {
  color: #ef4444 !important;
}

.settings-desc {
  font-size: 14px;
  color: var(--pm-text-secondary);
  margin: 0;
}

.settings-body {
  width: 100%;
}

.premium-input :deep(.el-input__wrapper) {
  box-shadow: none;
  border-radius: var(--pm-radius-sm);
  border: 1px solid var(--pm-border-color);
  background-color: var(--pm-bg-secondary);
  transition: all 0.2s;
}

.premium-input :deep(.el-input__wrapper:hover),
.premium-input :deep(.el-input__wrapper.is-focus) {
  border-color: var(--pm-primary-color);
  background-color: var(--pm-bg-primary);
}

.form-actions {
  margin-top: 24px;
}

.submit-btn {
  border-radius: var(--pm-radius-sm);
  font-weight: 500;
  padding: 10px 24px;
}

.logout-btn {
  border-radius: var(--pm-radius-sm);
  font-weight: 500;
}

.settings-divider {
  margin: 32px 0;
  border-color: var(--pm-border-color);
}
</style>