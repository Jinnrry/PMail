<template>
    <div id="main">
        <div id="title">{{ detailData.subject }}</div>
        <el-divider />

        <div>
            <span>{{ lang.to }}：
                <span class="userItem" v-for="to in tos">{{ to.Name }} {{ to.EmailAddress }} ;</span>
            </span>

            <span v-if="showCC">{{ lang.cc }}：
                <span class="userItem" v-for="ccs in cc">{{ cc.Name }} {{ cc.EmailAddress }} ;</span>
            </span>
        </div>
        <el-divider />
        <div class="content" id="text" v-if="detailData.html == ''">
            {{ detailData.text }}
        </div>

        <div class="content" id="html" v-else v-html="detailData.html">

        </div>

        <div v-if="detailData.attachments.length > 0" style="">
            <el-divider />
            {{ lang.attachment }}：
            <a class="att" v-for="item in detailData.attachments"
                :href="'/attachments/download/' + detailData.id + '/' + item.Index"> <el-icon>
                    <Document />
                </el-icon> {{ item.Filename }} </a>
        </div>


    </div>
</template>

<script setup>
import { RouterLink } from 'vue-router'
import $http from "../http/http";
import { reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import router from "@/router";  //根路由对象
import { Document } from '@element-plus/icons-vue';
import lang from '../i18n/i18n';

const route = useRoute()
const detailData = ref({
    attachments:[]
})

const tos = ref()
const ccs = ref()
const showCC = ref(false)

$http.post("/api/email/detail", { id: parseInt(route.params.id) }).then(res => {
    detailData.value = res.data
    if (res.data.to != "" && res.data.to != null) {
        tos.value = JSON.parse(res.data.to)
    }
    if (res.data.cc != "" && res.data.cc != null) {
        ccs.value = JSON.parse(res.data.cc)

    }

    if (ccs.value != null && ccs.value != undefined){
         showCC.value = ccs.value.length > 0
    }else{
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

#userItem {}

.content {
    /* background-color: aliceblue; */
}

a,a:link,a:visited,a:hover,a:active{
    text-decoration: none;
    color:inherit;
}

.att{
    display:block;
}
</style>