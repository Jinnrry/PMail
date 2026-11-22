<template>
  <div class="settings-card">
    <div class="settings-header">
      <h3>{{ lang.dns_records }}</h3>
      <p class="settings-desc">{{ lang.dns_desc }}</p>
    </div>
    <div v-for="(records, domain) in dnsInfos" :key="domain" class="domain-block">
      <h4>{{ domain }}</h4>
      <DNSRecordTable :records="records"/>
    </div>
  </div>
</template>

<script setup>
import {onMounted, ref} from 'vue'
import {ElNotification} from 'element-plus'
import lang from '../i18n/i18n';
import {http} from "@/utils/axios";
import DNSRecordTable from '@/components/DNSRecordTable.vue'

const dnsInfos = ref({})

onMounted(() => {
  http.post("/api/domain/dns", {}).then(res => {
    if (res.errorNo !== 0) {
      ElNotification({title: lang.fail, message: res.errorMsg, type: 'error'})
      return
    }
    dnsInfos.value = res.data
  })
})
</script>

<style scoped>
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

.domain-block {
  margin-bottom: 24px;
}

.domain-block h4 {
  font-size: 15px;
  font-weight: 600;
  color: var(--pm-text-primary);
  margin: 0 0 10px 0;
}
</style>
