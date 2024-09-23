<template>
  <el-table :data="data" :show-header="true">
    <el-table-column prop="id" label="id"/>
    <el-table-column prop="name" :label="lang.rule_name"/>
    <el-table-column prop="action" :label="lang.rule_do">
      <template #default="scope">
        {{ ActionName[scope.row.action] }}
      </template>
    </el-table-column>
    <el-table-column prop="params" :label="lang.rule_params"/>
    <el-table-column prop="sort" :label="lang.rule_priority"/>
    <el-table-column>
      <template #default="scope">
        <div style="display: flex; align-items: center">
          <el-button size="small" type="primary" :icon="Edit" circle @click="editRule(scope.row)"/>
          <el-popconfirm confirm-button-text="Yes" cancel-button-text="No, Thanks" :icon="InfoFilled"
                         @confirm="delRule(scope.row.id)" icon-color="#626AEF" :title="lang.del_rule_confirm">
            <template #reference>
              <el-button size="small" type="danger" :icon="Delete" circle/>
            </template>
          </el-popconfirm>
        </div>
      </template>
    </el-table-column>
  </el-table>

  <div>
    <el-button @click="dialogVisible = true">{{ lang.new_rule }}</el-button>
  </div>


  <el-dialog v-model="dialogVisible" :title="lang.new_rule" width="60%">
    <div style="text-align: left; padding-left: 20px;">
      <el-form v-model="addRuleForm" :inline="true" label-position="top">
        <el-form-item style="width: 400px;" :label="lang.rule_name">
          <el-input v-model="addRuleForm.name"/>
        </el-form-item>

        <el-form-item :label="lang.rule_priority">
          <el-input v-model="addRuleForm.sort" type="number" oninput="value=value.replace(/[^\-\d]/g, '')"/>
        </el-form-item>
        <el-divider/>
        <div style="width: 100%;">{{ lang.rule_desc }}</div>
        <div style="width: 100%;">
          <div v-for="(rule, index) in addRuleForm.rules" :key="index">
            <el-select v-model="rule.field" placeholder="Select">
              <el-option key="From" :label="lang.from" value="From"/>
              <el-option key="Subject" :label="lang.subject" value="Subject"/>
              <el-option key="To" :label="lang.to" value="To"/>
              <el-option key="Cc" :label="lang.cc" value="Cc"/>
              <el-option key="Content" :label="lang.content" value="Content"/>
            </el-select>

            <el-select v-model="rule.type" placeholder="Select">
              <el-option key="equal" :label="lang.equal" value="equal"/>
              <el-option key="contains" :label="lang.contains" value="contains"/>
              <el-option key="regex" :label="lang.regex" value="regex"/>
            </el-select>

            <el-input v-model="rule.rule" style="width: 350px;"/>
            <el-button size="small" type="danger" :icon="Delete" @click="removeRuleLine(index)" circle/>
          </div>
        </div>
        <div style="padding-top: 7px;">
          <el-button size="small" type="primary" :icon="Plus" circle @click="addRule()"/>
        </div>
        <el-divider/>
        <div style="width: 100%;">{{ lang.rule_do }}</div>
        <el-form-item>
          <el-select v-model="addRuleForm.action" placeholder="Select" @change="ruleTypeChange()">
            <el-option key="mark_read" :label="lang.mark_read" :value="READ"/>
            <el-option key="move" :label="lang.move" :value="MOVE"/>
            <el-option key="delete" :label="lang.delete" :value="DELETE"/>
            <el-option key="forward" :label="lang.forward" :value="FORWARD"/>
          </el-select>
          <el-select v-if="addRuleForm.action === 4" v-model="addRuleForm.params" @click="reflushGroupInfos">
            <el-option v-for="gp in groupData.list" :key="gp.id" :label="gp.name" :value="gp.id"/>
          </el-select>

          <el-input v-if="addRuleForm.action === 2" v-model="addRuleForm.params" style="width: 250px;"
                    placeholder="Forward Email Address"/>

        </el-form-item>

      </el-form>
    </div>
    <template #footer>
            <span class="dialog-footer">
                <el-button type="primary" @click="submitRule()">
                    {{ lang.submit }}
                </el-button>
            </span>
    </template>
  </el-dialog>
</template>

<script setup>
import {reactive, ref} from 'vue';
import lang from '../i18n/i18n';
import {Delete, Edit, InfoFilled, Plus} from '@element-plus/icons-vue'
import {http} from "@/utils/axios";
import {ElNotification} from "element-plus";

const data = ref([])
const dialogVisible = ref(false)
const READ = 1
const FORWARD = 2
const DELETE = 3
const MOVE = 4

const ActionName = {
  1: lang.mark_read,
  2: lang.forward,
  3: lang.delete,
  4: lang.move
}

const init = function () {
  http.post("/api/rule/get").then((res) => {
    data.value = res.data
  })
}

init()

const groupData = reactive({
  list: []
})

const reflushGroupInfos = function () {
  http.get('/api/group/list').then(res => {
    if (res.data != null) {
      groupData.list = res.data
      for (let i = 0; i < groupData.list.length; i++) {
        groupData.list[i].id += ""
      }
    }

  })
}

reflushGroupInfos()

const addRuleForm = reactive({
  "id": 0,
  "name": "",
  "sort": 0,
  "rules": [
    {
      "field": "",
      "type": "",
      "rule": ""
    }
  ],
  "action": "",
  "params": ""
})

const delRule = function (id) {
  http.post("/api/rule/del", {"id": id}).then((res) => {
    ElNotification({
      title: res.errorNo === 0 ? lang.succ : lang.fail,
      message: res.data,
      type: res.errorNo === 0 ? 'success' : 'error',
    })

    init()
  })
}

const editRule = function (ruleInfo) {
  addRuleForm.id = ruleInfo.id
  addRuleForm.name = ruleInfo.name
  addRuleForm.rules = ruleInfo.rules
  addRuleForm.action = ruleInfo.action
  addRuleForm.params = ruleInfo.params
  addRuleForm.sort = ruleInfo.sort
  dialogVisible.value = true
}

const removeRuleLine = function (index) {
  addRuleForm.rules.splice(index, 1);
}

const addRule = function () {
  addRuleForm.rules.push(
      {
        "field": "",
        "type": "",
        "rule": ""
      }
  )
}

const submitRule = function () {
  let api = "/api/rule/add"
  if (addRuleForm.id > 0) {
    api = "/api/rule/update"
  }

  addRuleForm.sort = parseInt(addRuleForm.sort)

  http.post(api, addRuleForm).then((res) => {
    if (res.errorNo !== 0) {
      ElNotification({
        title: lang.fail,
        message: res.data,
        type: 'error',
      })
    } else {
      init()
      dialogVisible.value = false

      addRuleForm.id = 0
      addRuleForm.name = ""
      addRuleForm.sort = 0
      addRuleForm.rules = [
        {
          "field": "",
          "type": "",
          "rule": ""
        }
      ]
      addRuleForm.action = ""
      addRuleForm.params = ""
    }
  })
}


const ruleTypeChange = function () {
  addRuleForm.params = ''
}
</script>
