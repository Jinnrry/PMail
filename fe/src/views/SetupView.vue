<template>
    <div id="main">
        <el-steps :active="active" align-center finish-status="success" id="status">
            <el-step :title="lang.welcome" />
            <el-step :title="lang.setDatabase" />
            <el-step :title="lang.SetDomail" />
            <el-step :title="lang.setDNS" />
            <el-step :title="lang.setSSL" />
            <el-step :title="lang.setOther" />
        </el-steps>


        <div v-if="active == 0" class="ctn">
            <div class="desc">
                <h2>{{ lang.tks_pmail }}</h2>
                <div style="margin-top: 10px;">{{ lang.guid_desc }}</div>
            </div>
        </div>




        <div v-if="active == 1" class="ctn">
            <div class="desc">
                <h2>{{ lang.select_db }}</h2>
                <div style="margin-top: 10px;">{{ lang.db_desc }}</div>
            </div>
            <div class="form" style="width: 400px;">
                <el-form label-width="120px">
                    <el-form-item :label="lang.type">
                        <el-select :placeholder="lang.db_select_ph" v-model="dbSettings.type">
                            <el-option label="MySQL" value="mysql" />
                            <el-option label="SQLite3" value="sqlite" />
                        </el-select>
                    </el-form-item>

                    <el-form-item :label="lang.mysql_dsn" v-if="dbSettings.type == 'mysql'">
                        <el-input :rows="2" type="textarea" v-model="dbSettings.dsn"
                            placeholder="root:12345@tcp(127.0.0.1:3306)/pmail?parseTime=True&loc=Local"></el-input>
                    </el-form-item>

                    <el-form-item :label="lang.sqlite_db_path" v-if="dbSettings.type == 'sqlite'">
                        <el-input v-model="dbSettings.dsn" placeholder="./pmail.db"></el-input>
                    </el-form-item>
                </el-form>
            </div>
        </div>


        <div v-if="active == 2" class="ctn">
            <div class="desc">
                <h2>{{ lang.SetDomail }}</h2>
                <!-- <div style="margin-top: 10px;">{{ lang.domain_desc }}</div> -->
            </div>
            <div class="form" style="width: 400px;">
                <el-form label-width="120px">

                    <el-form-item :label="lang.smtp_domain">
                        <el-input placeholder="domaim.com" v-model="domainSettings.smtp_domain">
                            <template #prepend>smtp.</template>
                        </el-input>
                    </el-form-item>

                    <el-form-item :label="lang.web_domain">
                        <el-input placeholder="pmail.domain.com" v-model="domainSettings.web_domain"></el-input>
                    </el-form-item>
                </el-form>
            </div>
        </div>


        <div v-if="active == 3" class="ctn_s">
            <div class="desc">
                <h2>{{ lang.setDNS }}</h2>
                <div style="margin-top: 10px;">{{ lang.dns_desc }}</div>
            </div>
            <div class="form" width="600px">
                <el-table :data="dnsInfos" border style="width: 100%">
                    <el-table-column prop="host" label="HOSTNAME" width="110px" />
                    <el-table-column prop="type" label="TYPE" width="110px" />
                    <el-table-column prop="value" label="VALUE">
                        <template #default="scope">
                            <div style="display: flex; align-items: center">
                                <el-tooltip :content="scope.row.tips" placement="top" v-if="scope.row.tips != ''">
                                    {{ scope.row.value }}
                                </el-tooltip>
                                <span v-else>{{ scope.row.value }}</span>
                            </div>
                        </template>

                    </el-table-column>
                    <el-table-column prop="ttl" label="TTL" width="110px" />
                </el-table>
            </div>
        </div>


        <el-button id="next" style="margin-top: 12px" @click="next">{{ lang.next }}</el-button>

    </div>
</template>

<script setup>
import $http from "../http/http";

import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import router from "@/router";  //根路由对象
import lang from '../i18n/i18n';


const dbSettings = reactive({
    "type": "",
    "dsn": "",
    "lable": ""
})

const domainSettings = reactive({
    "web_domain": "",
    "smtp_domain": ""
})

const active = ref(0)

const dnsInfos = ref([
    { "host": "smtp", "type": "A", "value": "YouServerIp", "prid": "NA", "ttl": "3600" }
])

const getDbConfig = () => {
    $http.post("/api/setup", { "action": "get", "step": "database" }).then((res) => {
        if (res.errorNo != 0) {
            ElMessage.error(res.errorMsg)
        } else {
            dbSettings.type = res.data.db_type;
            dbSettings.dsn = res.data.db_dsn;
        }
    })
}

const getDomainConfig = () => {
    $http.post("/api/setup", { "action": "get", "step": "domain" }).then((res) => {
        if (res.errorNo != 0) {
            ElMessage.error(res.errorMsg)
        } else {
            domainSettings.web_domain = res.data.web_domain;
            domainSettings.smtp_domain = res.data.smtp_domain;
        }
    })
}

const setDbConfig = () => {
    $http.post("/api/setup", { "action": "set", "step": "database", "db_type": dbSettings.type, "db_dsn": dbSettings.dsn }).then((res) => {
        if (res.errorNo != 0) {
            ElMessage.error(res.errorMsg)
        } else {
            active.value++;
            getDomainConfig();
        }
    })
}

const getDNSConfig = () => {
    $http.post("/api/setup", { "action": "get", "step": "dns" }).then((res) => {
        if (res.errorNo != 0) {
            ElMessage.error(res.errorMsg)
        } else {
            dnsInfos.value = res.data
        }
    })
}

const setDomainConfig = () => {
    $http.post("/api/setup", { "action": "set", "step": "domain", "web_domain": domainSettings.web_domain, "smtp_domain": domainSettings.smtp_domain }).then((res) => {
        if (res.errorNo != 0) {
            ElMessage.error(res.errorMsg)
        } else {
            active.value++;
            getDNSConfig();
        }
    })
}


const next = () => {
    switch (active.value) {
        case 0:
            active.value++
            getDbConfig();
            break
        case 1:
            setDbConfig();
            break;
        case 2:
            setDomainConfig();
            break;
        case 3:
            active.value++
            break
    }

}
</script>


<style scoped>
#main {
    width: 100%;
    height: 100%;
    background-color: #f1f1f1;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
}

.desc {
    padding-right: 20px;
}

#status {}

.ctn {
    display: flex;
    justify-content: center;
}

.ctn_s {
    display: flex;
    flex-direction: column;

}

#next {}
</style>