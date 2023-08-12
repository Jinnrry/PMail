<template>
    <div style="height: 100%">
        <div id="operation">
            <div id="action">
                <RouterLink to="/editer">+{{ lang.compose }}</RouterLink>
            </div>
            <!-- <div id="action">全部标记为已读</div> -->
        </div>
        <div id="title">{{ groupStore.name }}</div>
        <div id="table">
            <el-table :data="data" :show-header="true" :border="false" @row-click="rowClick" :row-style="rowStyle">
                <el-table-column type="selection" width="30" />
                <el-table-column prop="title" label="" width="50">
                    <template #default="scope">
                        <div>
                            <span v-if="!scope.row.is_read">
                                {{ lang.new }}
                            </span>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column prop="title" :label="lang.sender" width="150">
                    <template #default="scope">
                        <span v-if="scope.row.is_read">
                            <div v-if="scope.row.sender.Name != ''">{{ scope.row.sender.Name }}</div>
                            {{ scope.row.sender.EmailAddress }}
                        </span>
                        <span v-else style="font-weight:bolder;">
                            <div v-if="scope.row.sender.Name != ''">{{ scope.row.sender.Name }}</div>
                            {{ scope.row.sender.EmailAddress }}
                        </span>
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
            <el-pagination background layout="prev, pager, next" :page-count="totalPage" />
        </div>
    </div>
</template>



<script setup>
import $http from "../http/http";

import { RouterLink } from 'vue-router'
import { reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import router from "@/router";  //根路由对象
import useGroupStore from '../stores/group'
import lang from '../i18n/i18n';

const groupStore = useGroupStore()

const route = useRoute()



let tag = groupStore.tag;

if (tag == "") {
    tag = '{"type":0,"status":-1}'
}


watch(groupStore, async (newV, oldV) => {
    tag = newV.tag;
    if (tag == "") {
        tag = '{"type":0,"status":-1}'
    }
    data.value = []
    $http.post("/api/email/list", { tag: tag, page_size: 10 }).then(res => {
        data.value = res.data.list
        totalPage.value = res.data.total_page
    })
})



const data = ref([])
const totalPage = ref(0)

$http.post("/api/email/list", { tag: tag, page_size: 10 }).then(res => {
    data.value = res.data.list
    totalPage.value = res.data.total_page
})

const rowClick = function (row, column, event) {
    router.push("/detail/" + row.id)
}

const rowStyle = function ({ row, rowIndwx }) {
    return { 'cursor': 'pointer' }
}

</script>


<style scoped>
#action {
    text-align: left;
    font-size: 20px;
    line-height: 40px;
    padding-left: 10px;
    margin-right: 5px;
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