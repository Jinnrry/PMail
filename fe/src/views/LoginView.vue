<template>
  <div class="login-wrapper">
    <div class="login-container">
      <div class="login-left">
        <div class="brand-info">
          <h1>PMail</h1>
        </div>
      </div>
      <div class="login-right">
        <div class="login-form-box">
          <h2>{{ lang.login }}</h2>
          <p class="subtitle">{{ lang.login_subtitle }}</p>
          
          <el-form :model="form" class="auth-form" @keyup.enter="onSubmit" label-position="top">
            <el-form-item :label="lang.account">
              <el-input 
                v-model="form.account" 
                :placeholder="lang.account"
                size="large"
                class="premium-input"
              />
            </el-form-item>
            <el-form-item :label="lang.password">
              <el-input 
                v-model="form.password" 
                :placeholder="lang.password"
                type="password" 
                size="large"
                class="premium-input"
                show-password
              />
            </el-form-item>
            <div class="submit-action">
              <el-button type="primary" @click="onSubmit" size="large" class="login-btn" :loading="loading">
                {{ lang.login }}
              </el-button>
            </div>
          </el-form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import {reactive, ref} from 'vue'
import {ElMessage} from 'element-plus'
import {router} from "@/router";
import lang from '../i18n/i18n';
import {http} from "@/utils/axios";
import {useGlobalStatusStore} from "@/stores/useGlobalStatusStore";

const globalStatus = useGlobalStatusStore();
const loading = ref(false);

const form = reactive({
  account: '',
  password: '',
})

const onSubmit = () => {
  if (!form.account || !form.password) {
    ElMessage.warning(lang.login_fill_required);
    return;
  }
  loading.value = true;
  http.post("/api/login", form).then(res => {
    loading.value = false;
    if (res.errorNo !== 0) {
      ElMessage.error(res.errorMsg)
    } else {
      Object.assign(globalStatus.userInfos , res.data) 
      router.replace({
        path: '/',
        query: {
          redirect: router.currentRoute.fullPath
        }
      })
    }
  }).catch(() => {
    loading.value = false;
  })
}
</script>

<style scoped>
.login-wrapper {
  width: 100vw;
  height: 100vh;
  background:
      radial-gradient(circle at 15% 20%, rgba(0, 113, 227, 0.1) 0%, transparent 36%),
      radial-gradient(circle at 85% 80%, rgba(0, 113, 227, 0.12) 0%, transparent 34%),
      var(--pm-bg-primary);
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 24px;
}

.login-container {
  width: 100%;
  max-width: 1080px;
  min-height: 640px;
  background: var(--pm-surface-glass);
  border-radius: var(--pm-radius-xl);
  box-shadow: var(--pm-shadow-lg);
  border: 1px solid var(--pm-glass-border);
  backdrop-filter: blur(18px);
  display: flex;
  overflow: hidden;
  animation: pm-rise-in 0.55s var(--pm-ease-out);
}

.login-left {
  flex: 1;
  background: linear-gradient(145deg, #0f6ddf 0%, #3f8ef7 55%, #9ec9ff 100%);
  padding: 72px 64px;
  color: #fff;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.brand-info h1 {
  font-size: 56px;
  font-weight: 600;
  margin-bottom: 14px;
  letter-spacing: -0.03em;
  animation: pm-rise-in 0.58s var(--pm-ease-out);
}

.login-right {
  flex: 1;
  padding: 64px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  background: linear-gradient(180deg, var(--pm-login-right-grad-start), var(--pm-login-right-grad-end));
}

.login-form-box {
  width: 100%;
  max-width: 400px;
  margin: 0 auto;
}

.login-form-box h2 {
  font-size: 34px;
  font-weight: 600;
  color: var(--pm-text-primary);
  margin-bottom: 8px;
  letter-spacing: -0.02em;
}

.subtitle {
  color: var(--pm-text-secondary);
  margin-bottom: 30px;
  font-size: 14px;
}

.auth-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: var(--pm-text-primary);
  padding-bottom: 4px;
}

.premium-input :deep(.el-input__wrapper) {
  background-color: var(--pm-surface-muted);
  border-radius: 12px;
  border: 1px solid transparent;
  box-shadow: 0 0 0 1px var(--pm-border-color) inset !important;
  transition: all 0.2s;
}

.premium-input :deep(.el-input__wrapper:hover),
.premium-input :deep(.el-input__wrapper.is-focus) {
  border-color: var(--pm-primary-color);
  background-color: var(--pm-bg-secondary);
}

.submit-action {
  margin-top: 34px;
}

.login-btn {
  width: 100%;
  font-weight: 600;
  border-radius: 12px;
  height: 50px;
  font-size: 15px;
  letter-spacing: 0.02em;
  box-shadow: 0 10px 24px rgba(0, 113, 227, 0.28);
  transition: transform 0.2s var(--pm-ease-out), box-shadow 0.2s var(--pm-ease-out);
}

.login-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 14px 28px rgba(0, 113, 227, 0.32);
}

@media (max-width: 768px) {
  .login-container {
    flex-direction: column;
    height: auto;
    min-height: 0;
    border-radius: var(--pm-radius-lg);
  }
  .login-left {
    padding: 38px 28px;
    align-items: center;
    text-align: center;
  }
  .brand-info h1 {
    font-size: 38px;
  }
  .login-right {
    padding: 36px 20px;
  }
}
</style>