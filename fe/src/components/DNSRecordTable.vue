<template>
  <el-table :data="rows" border style="width: 100%">
    <el-table-column prop="host" label="HOSTNAME" width="110px">
      <template #default="scope">
        <div style="display: flex; align-items: center">
          <el-tooltip :content="lang.dns_root_desc" placement="top"
                      v-if="scope.row.host === '' || scope.row.host === '@' ">
            {{ scope.row.host }}
          </el-tooltip>
          <span v-else>{{ scope.row.host }}</span>
        </div>
      </template>
    </el-table-column>
    <el-table-column prop="type" label="TYPE" width="110px"/>
    <el-table-column prop="value" label="VALUE">
      <template #default="scope">
        <div style="display: flex; align-items: center; word-break: break-all">
          <el-tooltip :content="scope.row.tips" placement="top" v-if="scope.row.tips !== ''">
            {{ scope.row.value }}
          </el-tooltip>
          <span v-else>{{ scope.row.value }}</span>
        </div>
        <div v-if="scope.row.splitValue" style="margin-top: 8px; word-break: break-all">
          <div style="font-size: 12px; color: var(--el-text-color-secondary)">{{ lang.dns_txt_split_desc }}</div>
          <div>{{ scope.row.splitValue }}</div>
        </div>
      </template>
    </el-table-column>
    <el-table-column prop="ttl" label="TTL" width="110px"/>
  </el-table>
</template>

<script setup>
import {computed} from 'vue'
import lang from '../i18n/i18n';

const props = defineProps({
  records: {type: Array, required: true}
})

// DNS 单个 TXT 字符串最长 255 字节，更长的值（如 2048 位 DKIM 公钥）拆成多段带引号的字符串
const splitTXT = (value) => value.match(/.{1,255}/g).map(s => `"${s}"`).join(' ')

const rows = computed(() => props.records.map(r => ({
  ...r,
  splitValue: r.type === 'TXT' && r.value.length > 255 ? splitTXT(r.value) : ''
})))
</script>
