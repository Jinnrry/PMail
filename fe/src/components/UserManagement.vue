<template>
  <div id="main">
    <el-table :data="userList" style="width: 100%">
      <el-table-column label="ID" prop="ID"/>
      <el-table-column :label="lang.account" prop="Account"/>
      <el-table-column :label="lang.user_name" prop="Name"/>
      <el-table-column :label="lang.disabled" prop="Disabled">
        <template #default="scope">
          <span>{{ scope.row.Disabled === 1 ? lang.disabled : lang.enabled }}</span>
        </template>
      </el-table-column>
      <el-table-column align="right">
        <template #header>
          <el-button type="primary" size="small" @click="createUser">
            New
          </el-button>
        </template>
        <template #default="scope">
          <el-button size="small" @click="handleEdit(scope.$index, scope.row)">
            Edit
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    <div id="paginationBox">
      <el-pagination v-model:current-page="currentPage" small background layout="prev, pager, next"
                     :page-count="totalPage" class="mt-4" @current-change="reflushList"/>
    </div>


    <el-dialog v-model="userInfoDialog" :title="title" width="500">
      <el-form>
        <el-form-item label-width="100px" :label="lang.account">
          <el-input :disabled="editModel === 'edit'" v-model="editUserInfo.account" autocomplete="off"/>
        </el-form-item>

        <el-form-item label-width="100px" :label="lang.user_name">
          <el-input v-model="editUserInfo.name" autocomplete="off"/>
        </el-form-item>

        <el-form-item label-width="100px" :label="lang.password">
          <el-input :placeholder="lang.resetPwd" v-model="editUserInfo.password" autocomplete="off"/>
        </el-form-item>

        <div style="display: flex;">
          <div
              style="display: inline-flex;justify-content: flex-end;align-items: flex-start;flex: 0 0 auto;font-size: var(--el-form-label-font-size); height: 32px;line-height: 32px;padding: 0 12px 0 60px;box-sizing: border-box; ">
            <el-switch v-model="editUserInfo.disabled" class="ml-2" :active-text="lang.disabled"
                       :inactive-text="lang.enabled"/>
          </div>


        </div>

      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="userInfoDialog = false">Cancel</el-button>
          <el-button type="primary" @click="submit">
            Confirm
          </el-button>
        </div>
      </template>
    </el-dialog>


  </div>
</template>


<script setup>
import {reactive, ref} from 'vue'
import lang from '../i18n/i18n';
import {http} from "@/utils/axios";
import {ElNotification} from "element-plus";

const userList = reactive([])
const currentPage = ref(1)
const totalPage = ref(1)
const userInfoDialog = ref(false)
const editModel = ref("edit")
const editUserInfo = reactive({
  "account": "",
  "name": "",
  "password": "",
  "disabled": false
})
const title = ref(lang.editUser)

const reflushList = function () {
  http.post('/api/user/list', {"current_page": currentPage.value, "page_size": 10}).then(res => {
    userList.length = 0

    totalPage.value = res.data.total_page
    userList.push(...res.data["list"])
  })
}


const handleEdit = function (idx, row) {
  editUserInfo.account = row.Account
  editUserInfo.name = row.Name
  editUserInfo.disabled = row.Disabled === 1
  editUserInfo.password = ""
  editModel.value = "edit"
  title.value = lang.editUser
  userInfoDialog.value = true
}

const createUser = function () {
  editUserInfo.account = ""
  editUserInfo.name = ""
  editUserInfo.disabled = false
  editUserInfo.password = ""
  editModel.value = "create"
  title.value = lang.newUser
  userInfoDialog.value = true
}


const submit = function () {
  if (editModel.value === 'edit') {

    let newData = {
      "account": editUserInfo.account,
      "username": editUserInfo.name,
      "disabled": editUserInfo.disabled ? 1 : 0
    }
    if (editUserInfo.password !== "") {
      newData["password"] = editUserInfo.password
    }

    http.post('/api/user/edit', newData).then(res => {
      ElNotification({
        title: res.errorNo === 0 ? lang.succ : lang.fail,
        message: res.errorNo === 0 ? "" : res.data,
        type: res.errorNo === 0 ? 'success' : 'error',
      })
      if (res.errorNo === 0) {
        reflushList()
        userInfoDialog.value = false
      }
    })
  } else {
    let newData = {
      "account": editUserInfo.account,
      "username": editUserInfo.name,
      "disabled": editUserInfo.disabled ? 1 : 0,
      "password": editUserInfo.password
    }

    http.post('/api/user/create', newData).then(res => {
      ElNotification({
        title: res.errorNo === 0 ? lang.succ : lang.fail,
        message: res.errorNo === 0 ? "" : res.data,
        type: res.errorNo === 0 ? 'success' : 'error',
      })
      if (res.errorNo === 0) {
        reflushList()
        userInfoDialog.value = false
      }
    })
  }
}


reflushList()


</script>


<style scoped>
#paginationBox {
  margin-top: 10px;
  display: flex;
  justify-content: center;
}
</style>