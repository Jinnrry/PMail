<template>
  <div id="main">
    <div id="title">{{ detailData.subject }}</div>
    <el-divider/>

    <div>
      <div>{{ lang.to }}：
        <el-tooltip v-for="to in tos" :key.prop="to" class="box-item" effect="dark" :content="to.EmailAddress" placement="top">
          <el-tag size="small" type="info">{{ to.Name !== '' ? to.Name : to.EmailAddress }}</el-tag>
        </el-tooltip>
      </div>

      <div v-if="showCC">{{ lang.cc }}：
        <el-tooltip v-for="item in ccs" :key="item" class="box-item" effect="dark" :content="item.EmailAddress" placement="top">
          <el-tag size="small" type="info">{{ item.Name !== '' ? item.Name : item.EmailAddress }}</el-tag>
        </el-tooltip>
      </div>

      <div>{{ lang.sender }}：
        <el-tooltip class="box-item" effect="dark" :content="detailData.from_address" placement="top">
          <el-tag size="small" type="info">
            {{ detailData.from_name !== '' ? detailData.from_name : detailData.from_address }}
          </el-tag>
        </el-tooltip>
      </div>

      <div>{{ lang.date }}：
        {{ detailData.send_date }}
      </div>
    </div>
    <el-divider/>
    <div class="content" id="text" v-if="detailData.html === ''">
      {{ detailData.text }}
    </div>

    <div class="content" id="html" v-else v-html="detailData.html">

    </div>

    <div v-if="detailData.attachments.length > 0" style="">
      <el-divider/>
      {{ lang.attachment }}：
      <a class="att" v-for="item in detailData.attachments" :key="item"
         :href="'/attachments/download/' + detailData.id + '/' + item.Index">
        <el-icon>
          <Document/>
        </el-icon>
        {{ item.Filename }} </a>
    </div>


  </div>
</template>

<script setup>

import {ref} from 'vue'
import {useRoute} from 'vue-router'
import {Document} from '@element-plus/icons-vue';
import lang from '../i18n/i18n';
import {http} from "@/utils/axios";

const route = useRoute()
const detailData = ref({
  attachments: []
})

const tos = ref()
const ccs = ref()
const showCC = ref(false)

http.post("/api/email/detail", {id: parseInt(route.params.id)}).then(res => {
  detailData.value = res.data
  if (res.data.to !== "" && res.data.to != null) {
    tos.value = JSON.parse(res.data.to)
  }
  if (res.data.cc !== "" && res.data.cc != null) {
    ccs.value = JSON.parse(res.data.cc)

  }

  if (ccs.value != null) {
    showCC.value = ccs.value.length > 0
  } else {
    showCC.value = false
  }
})
</script>

<style scoped>
#main {
  display: flex;
  padding-left: 20px;
  padding-right: 80px;
  text-align: left;
}

#title {
  font-size: 40px;
  text-align: left;
}

#userItem {
}

.content {
  /* background-color: aliceblue; */
}

a, a:link, a:visited, a:hover, a:active {
  text-decoration: none;
  color: inherit;
}

.att {
  display: block;
}
</style>