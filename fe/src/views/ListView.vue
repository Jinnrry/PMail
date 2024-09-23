<template>
  <div style="height: 100%">
    <div id="operation">
      <div id="action">
        <RouterLink to="/editer">+{{ lang.compose }}</RouterLink>
      </div>
    </div>
    <div id="title">{{ groupStore.name }}</div>
    <div id="action">
      <el-button @click="del" size="small">{{ lang.del_btn }}</el-button>
      <el-button @click="markRead" size="small">{{ lang.read_btn }}</el-button>
      <el-dropdown style="margin-left: 12px;">
        <el-button size="small">
          {{ lang.move_btn }}
          <el-icon class="el-icon--right">
            <EpArrowDownBold />
          </el-icon>
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item @click="move(group.id)" v-for="group in groupList" :key="group.id">{{
                group.name
              }}
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
    <div id="table">
      <el-table ref="taskTableDataRef" :data="data" :show-header="true"
                :border="false" @row-click="rowClick" :row-style="rowStyle">
        <el-table-column type="selection" width="30"/>
        <el-table-column prop="is_read" label="" width="50">
          <template #default="scope">
            <div>
                            <span v-if="!scope.row.is_read">
                                {{ lang.new }}
                            </span>
              <span style="font-weight: 900;color: #FF0000;" v-if="scope.row.dangerous">
                                <el-tooltip effect="dark" :content="lang.dangerous" placement="top-start">
                                    !
                                </el-tooltip>

                            </span>
              <span style="font-weight: 900;color: #FF0000;" v-if="scope.row.error !== ''">
                                <el-tooltip effect="dark" :content="scope.row.error" placement="top-start">
                                    !
                                </el-tooltip>

                            </span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="title" :label="lang.sender" width="150">
          <template #default="scope">
            <el-tooltip class="box-item" effect="dark" :content="scope.row.sender.EmailAddress" placement="top">
              <el-tag size="small" type="info">
                {{ scope.row.sender.Name !== '' ? scope.row.sender.Name : scope.row.sender.EmailAddress }}
              </el-tag>
            </el-tooltip>
          </template>
        </el-table-column>

        <el-table-column prop="title" :label="lang.to" width="150">
          <template #default="scope">
            <el-tooltip v-for="toInfo in scope.row.to" :key="toInfo" class="box-item" effect="dark" :content="toInfo.EmailAddress"
                        placement="top">
              <el-tag size="small" type="info">{{ toInfo.Name !== '' ? toInfo.Name : toInfo.EmailAddress }}</el-tag>
            </el-tooltip>
          </template>
        </el-table-column>

        <el-table-column prop="desc" :label="lang.title">
          <template #default="scope">
            <div v-if="scope.row.is_read">{{ scope.row.title }}</div>
            <div v-else style="font-weight:bolder;">{{ scope.row.title }}</div>

            <div style="font-size: 12px;height: 24px;">{{ scope.row.desc }}</div>

          </template>
        </el-table-column>
        <el-table-column prop="datetime" :label="lang.date" width="180">
          <template #default="scope">
            <span v-if="scope.row.is_read">{{ scope.row.datetime }}</span>
            <span v-else style="font-weight:bolder;">{{ scope.row.datetime }}</span>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <div id="pagination">
      <el-pagination background layout="prev, pager, next" :page-count="totalPage" @current-change="pageChange"/>
    </div>
  </div>
</template>


<script setup>

import {EpArrowDownBold} from "vue-icons-plus/ep";
import {RouterLink, useRouter} from 'vue-router'
import {ref, watch} from 'vue'
import useGroupStore from '../stores/group'
import lang from '../i18n/i18n';
import {http} from "@/utils/axios";
import {ElMessage, ElMessageBox} from "element-plus";


const router = useRouter();
const groupStore = useGroupStore()
const groupList = ref([])
const taskTableDataRef = ref(null)
let tag = groupStore.tag;

if (tag === "") {
  tag = '{"type":0,"status":-1}'
}


watch(groupStore, async (newV) => {
  tag = newV.tag;
  if (tag === "") {
    tag = '{"type":0,"status":-1}'
  }
  data.value = []
  http.post("/api/email/list", {tag: tag, page_size: 10}).then(res => {
    data.value = res.data.list
    totalPage.value = res.data.total_page
  })
})


const data = ref([])
const totalPage = ref(0)

const updateList = function () {
  http.post("/api/email/list", {tag: tag, page_size: 10}).then(res => {
    data.value = res.data.list
    totalPage.value = res.data.total_page
  })
}

const updateGroupList = function () {
  http.post("/api/group/list").then(res => {
    groupList.value = res.data
  })
}

updateList()
updateGroupList()

const rowClick = function (row) {
  router.push("/detail/" + row.id)
}

const markRead = function () {
  let rows = taskTableDataRef.value?.getSelectionRows()
  let ids = []
  rows.forEach(element => {
    ids.push(element.id)
  });

  http.post("/api/email/read", {"ids": ids}).then(res => {
    if (res.errorNo === 0) {
      updateList()
    } else {
      ElMessage({
        type: 'error',
        message: res.errorMsg,
      })
    }
  })
}


const move = function (group_id) {
  let rows = taskTableDataRef.value?.getSelectionRows()
  let ids = []
  rows.forEach(element => {
    ids.push(element.id)
  });

  ElMessageBox.confirm(
      lang.move_email_confirm,
      'Warning',
      {
        confirmButtonText: 'OK',
        cancelButtonText: 'Cancel',
        type: 'warning',
      }
  )
      .then(() => {
        http.post("/api/email/move", {"group_id": group_id, "ids": ids}).then(res => {
          if (res.errorNo === 0) {
            updateList()
            ElMessage({
              type: 'success',
              message: 'Move completed',
            })
          } else {
            ElMessage({
              type: 'error',
              message: res.errorMsg,
            })
          }
        })


      })
}


const del = function () {
  let rows = taskTableDataRef.value?.getSelectionRows()
  let ids = []
  rows.forEach(element => {
    ids.push(element.id)
  });

  let groupTag = JSON.parse(tag)


  ElMessageBox.confirm(
      lang.del_email_confirm,
      'Warning',
      {
        confirmButtonText: 'OK',
        cancelButtonText: 'Cancel',
        type: 'warning',
      }
  )
      .then(() => {
        http.post("/api/email/del", {"ids": ids, "forcedDel": groupTag.status === 3}).then(res => {
          if (res.errorNo === 0) {
            updateList()
            ElMessage({
              type: 'success',
              message: 'Delete completed',
            })
          } else {
            ElMessage({
              type: 'error',
              message: res.errorMsg,
            })
          }
        })


      })
}


const rowStyle = function () {
  return {'cursor': 'pointer'}
}

const pageChange = function (p) {
  http.post("/api/email/list", {tag: tag, page_size: 10, current_page: p}).then(res => {
    data.value = res.data.list
  })
}

</script>


<style scoped>
#action {
  display: flex;
  flex-direction: row;
}


#action a,
a:visited {
  color: #000000;
  text-decoration: none;
}

#operation {
  display: flex;
  height: 40px;
  background-color: rgb(236, 244, 251);
}

#title {
  margin-top: 10px;
  font-size: 23px;
  text-align: left;
  padding-left: 20px;
}

#table {
  text-align: left;
  width: 100%;
  padding-left: 20px;
}

#pagination {
  padding-top: 30px;
  display: flex;
  justify-content: center;
  /* 水平居中 */
  width: 100%;
}
</style>