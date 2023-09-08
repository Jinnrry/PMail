<template>
    <div id="main">
        <el-steps :active="active" align-center finish-status="success" id="status">
            <el-step :title="lang.welcome" />
            <el-step :title="lang.setDatabase" />
            <el-step :title="lang.setAdminPassword" />
            <el-step :title="lang.SetDomail" />
            <el-step :title="lang.setDNS" />
            <el-step :title="lang.setSSL" />
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
                        <el-select :placeholder="lang.db_select_ph" v-model="dbSettings.type" @change="dbSettings.dsn=''">
                            <el-option label="MySQL" value="mysql" />
                            <el-option label="SQLite3" value="sqlite" />
                        </el-select>
                    </el-form-item>

                    <el-form-item :label="lang.mysql_dsn" v-if="dbSettings.type == 'mysql'">
                        <el-input :rows="2" type="textarea" v-model="dbSettings.dsn"
                            placeholder="root:12345@tcp(127.0.0.1:3306)/pmail?parseTime=True&loc=Local"></el-input>
                    </el-form-item>

                    <el-form-item :label="lang.sqlite_db_path" v-if="dbSettings.type == 'sqlite'">
                        <el-input v-model="dbSettings.dsn" placeholder="./config/pmail.db"></el-input>
                    </el-form-item>
                </el-form>
            </div>
        </div>


        <div v-if="active == 2" class="ctn">
            <div class="desc">
                <h2>{{ lang.setAdminPassword }}</h2>
                <!-- <div style="margin-top: 10px;">{{ lang.domain_desc }}</div> -->
            </div>
            <div class="form" style="width: 400px;">
                <el-form label-width="120px">

                    <el-form-item :label="lang.admin_account">
                        <el-input v-bind:disabled="adminSettings.hadSeted" placeholder="admin"
                            v-model="adminSettings.account"></el-input>
                    </el-form-item>

                    <el-form-item :label="lang.password">
                        <el-input type="password" v-bind:disabled="adminSettings.hadSeted" placeholder=""
                            v-model="adminSettings.password"></el-input>
                    </el-form-item>

                    <el-form-item :label="lang.enter_again">
                        <el-input type="password" v-bind:disabled="adminSettings.hadSeted" placeholder=""
                            v-model="adminSettings.password2"></el-input>
                    </el-form-item>
                </el-form>
            </div>
        </div>


        <div v-if="active == 3" class="ctn">
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


        <div v-if="active == 4" class="ctn_s">
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

        <div v-if="active == 5" class="ctn">
            <div class="desc">
                <h2>{{ lang.setSSL }}</h2>
                <div style="margin-top: 10px;">{{ lang.setSSL }}</div>
            </div>
            <div class="form" width="600px">
                <el-form label-width="120px">
                    <el-form-item :label="lang.type">
                        <el-select :placeholder="lang.ssl_auto" v-model="sslSettings.type">
                            <el-option :label="lang.ssl_auto" value="0" />
                            <el-option :label="lang.ssl_manuallyf" value="1" />
                        </el-select>
                    </el-form-item>

                    <el-form-item :label="lang.ssl_key_path" v-if="sslSettings.type == '1'">
                        <el-input placeholder="./config/ssl/private.key" v-model="sslSettings.key_path"></el-input>
                    </el-form-item>

                    <el-form-item :label="lang.ssl_crt_path" v-if="sslSettings.type == '1'">
                        <el-input placeholder="./config/ssl/public.crt" v-model="sslSettings.crt_path"></el-input>
                    </el-form-item>

                </el-form>
            </div>
        </div>

        <el-button v-loading.fullscreen.lock="fullscreenLoading" id="next" style="margin-top: 12px" @click="next">{{
            lang.next }}</el-button>

    </div>
</template>

<script setup>
import $http from "../http/http";

import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import router from "@/router";  //根路由对象
import lang from '../i18n/i18n';

const adminSettings = reactive({
    "account": "admin",
    "password": "",
    "password2": "",
    "hadSeted": false
})

const dbSettings = reactive({
    "type": "sqlite",
    "dsn": "./config/pmail.db",
    "lable": ""
})

const domainSettings = reactive({
    "web_domain": "",
    "smtp_domain": ""
})

const sslSettings = reactive({
    "type": "0",
    "key_path": "./config/ssl/private.key",
    "crt_path": "./config/ssl/public.crt"
})

const active = ref(0)
const fullscreenLoading = ref(false)


const dnsInfos = ref([
])

const setPassword = () => {
    if (adminSettings.hadSeted) {
        active.value++;
        getDomainConfig();
        return;
    }

    if (adminSettings.password != adminSettings.password2) {
        ElMessage.error(lang.err_pwd_diff)
    } else {
        $http.post("/api/setup", { "action": "set", "step": "password", "account": adminSettings.account, "password": adminSettings.password }).then((res) => {
            if (res.errorNo != 0) {
                ElMessage.error(res.errorMsg)
            } else {
                active.value++;
                getDomainConfig();
            }
        })
    }
}

const getPassword = () => {
    $http.post("/api/setup", { "action": "get", "step": "password" }).then((res) => {
        if (res.errorNo != 0) {
            ElMessage.error(res.errorMsg)
        } else {
            adminSettings.hadSeted = res.data != ""
            if (adminSettings.hadSeted) {
                adminSettings.account = res.data
                adminSettings.password = "*******"
                adminSettings.password2 = "*******"
            }

        }
    })
}


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
            getPassword();
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


const getSSLConfig = () => {
    $http.post("/api/setup", { "action": "get", "step": "ssl" }).then((res) => {
        if (res.errorNo != 0) {
            ElMessage.error(res.errorMsg)
        } else {
            sslSettings.type = res.data
        }
    })
}


const setSSLConfig = () => {
    fullscreenLoading.value = true;
    $http.post("/api/setup", { "action": "set", "step": "ssl", "ssl_type": sslSettings.type, "key_path": sslSettings.key_path, "crt_path": sslSettings.crt_path }).then((res) => {
        if (res.errorNo != 0) {
            fullscreenLoading.value = false;
            ElMessage.error(res.errorMsg)
        } else {
            setTimeout(function () {
                window.location.href = "https://" + domainSettings.web_domain;
            }, 10000);

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
            setPassword();
            break;
        case 3:
            setDomainConfig();
            break;
        case 4:
            getSSLConfig();
            active.value++
            break
        case 5:
            setSSLConfig();
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