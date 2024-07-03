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
                        <el-select :placeholder="lang.db_select_ph" v-model="dbSettings.type"
                            @change="dbSettings.dsn = ''">
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

        <el-alert :closable="false" title="Warning!" type="error" center
            v-if="active == 5 && sslSettings.type == 0 && port != 80" :description="lang.autoSSLWarn" />

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

                    <el-form-item :label="lang.ssl_challenge_type" v-if="sslSettings.type == '0'">
                        <el-select :placeholder="lang.ssl_auto_http" v-model="sslSettings.challenge">
                            <el-option :label="lang.ssl_auto_http" value="http" />
                            <!-- <el-option :label="lang.ssl_auto_dns" value="dns" /> -->
                        </el-select>

                        <el-tooltip class="box-item" effect="dark" :content="lang.challenge_typ_desc"
                            placement="top-start">
                            <span style="margin-left: 6px; font-size:18px; font-weight: bolder;">?</span>
                        </el-tooltip>
                    </el-form-item>

                    <el-form-item :label="lang.oomain_service_provider"
                        v-if="sslSettings.type == '0' && sslSettings.challenge == 'dns'">
                        <el-select @change="provide_change" :placeholder="lang.oomain_service_provider"
                            v-model="sslSettings.provider">
                            <el-option label="acme-dns" value="acme-dns" />
                            <el-option label="alidns" value="alidns" />
                            <el-option label="allinkl" value="allinkl" />
                            <el-option label="arvancloud" value="arvancloud" />
                            <el-option label="azure" value="azure" />
                            <el-option label="azuredns" value="azuredns" />
                            <el-option label="auroradns" value="auroradns" />
                            <el-option label="autodns" value="autodns" />
                            <el-option label="bindman" value="bindman" />
                            <el-option label="bluecat" value="bluecat" />
                            <el-option label="brandit" value="brandit" />
                            <el-option label="bunny" value="bunny" />
                            <el-option label="checkdomain" value="checkdomain" />
                            <el-option label="civo" value="civo" />
                            <el-option label="clouddns" value="clouddns" />
                            <el-option label="cloudflare" value="cloudflare" />
                            <el-option label="cloudns" value="cloudns" />
                            <el-option label="cloudru" value="cloudru" />
                            <el-option label="cloudxns" value="cloudxns" />
                            <el-option label="conoha" value="conoha" />
                            <el-option label="constellix" value="constellix" />
                            <el-option label="cpanel" value="cpanel" />
                            <el-option label="derak" value="derak" />
                            <el-option label="desec" value="desec" />
                            <el-option label="designate" value="designate" />
                            <el-option label="digitalocean" value="digitalocean" />
                            <el-option label="dnshomede" value="dnshomede" />
                            <el-option label="dnsimple" value="dnsimple" />
                            <el-option label="dnsmadeeasy" value="dnsmadeeasy" />
                            <el-option label="dnspod" value="dnspod" />
                            <el-option label="dode" value="dode" />
                            <el-option label="domeneshop" value="domeneshop" />
                            <el-option label="domainnameshop" value="domainnameshop" />
                            <el-option label="dreamhost" value="dreamhost" />
                            <el-option label="duckdns" value="duckdns" />
                            <el-option label="dyn" value="dyn" />
                            <el-option label="dynu" value="dynu" />
                            <el-option label="easydns" value="easydns" />
                            <el-option label="edgedns" value="edgedns" />
                            <el-option label="fastdns" value="fastdns" />
                            <el-option label="efficientip" value="efficientip" />
                            <el-option label="epik" value="epik" />
                            <el-option label="exec" value="exec" />
                            <el-option label="exoscale" value="exoscale" />
                            <el-option label="freemyip" value="freemyip" />
                            <el-option label="gandi" value="gandi" />
                            <el-option label="gandiv5" value="gandiv5" />
                            <el-option label="gcloud" value="gcloud" />
                            <el-option label="gcore" value="gcore" />
                            <el-option label="glesys" value="glesys" />
                            <el-option label="godaddy" value="godaddy" />
                            <el-option label="googledomains" value="googledomains" />
                            <el-option label="hetzner" value="hetzner" />
                            <el-option label="hostingde" value="hostingde" />
                            <el-option label="hosttech" value="hosttech" />
                            <el-option label="httpreq" value="httpreq" />
                            <el-option label="hurricane" value="hurricane" />
                            <el-option label="hyperone" value="hyperone" />
                            <el-option label="ibmcloud" value="ibmcloud" />
                            <el-option label="iij" value="iij" />
                            <el-option label="iijdpf" value="iijdpf" />
                            <el-option label="infoblox" value="infoblox" />
                            <el-option label="infomaniak" value="infomaniak" />
                            <el-option label="internetbs" value="internetbs" />
                            <el-option label="inwx" value="inwx" />
                            <el-option label="ionos" value="ionos" />
                            <el-option label="ipv64" value="ipv64" />
                            <el-option label="iwantmyname" value="iwantmyname" />
                            <el-option label="joker" value="joker" />
                            <el-option label="liara" value="liara" />
                            <el-option label="lightsail" value="lightsail" />
                            <el-option label="linode" value="linode" />
                            <el-option label="linodev4" value="linodev4" />
                            <el-option label="liquidweb" value="liquidweb" />
                            <el-option label="loopia" value="loopia" />
                            <el-option label="luadns" value="luadns" />
                            <el-option label="mailinabox" value="mailinabox" />
                            <el-option label="manual" value="manual" />
                            <el-option label="metaname" value="metaname" />
                            <el-option label="mydnsjp" value="mydnsjp" />
                            <el-option label="mythicbeasts" value="mythicbeasts" />
                            <el-option label="namecheap" value="namecheap" />
                            <el-option label="namedotcom" value="namedotcom" />
                            <el-option label="namesilo" value="namesilo" />
                            <el-option label="nearlyfreespeech" value="nearlyfreespeech" />
                            <el-option label="netcup" value="netcup" />
                            <el-option label="netlify" value="netlify" />
                            <el-option label="nicmanager" value="nicmanager" />
                            <el-option label="nifcloud" value="nifcloud" />
                            <el-option label="njalla" value="njalla" />
                            <el-option label="nodion" value="nodion" />
                            <el-option label="ns1" value="ns1" />
                            <el-option label="oraclecloud" value="oraclecloud" />
                            <el-option label="otc" value="otc" />
                            <el-option label="ovh" value="ovh" />
                            <el-option label="pdns" value="pdns" />
                            <el-option label="plesk" value="plesk" />
                            <el-option label="porkbun" value="porkbun" />
                            <el-option label="rackspace" value="rackspace" />
                            <el-option label="rcodezero" value="rcodezero" />
                            <el-option label="regru" value="regru" />
                            <el-option label="rfc2136" value="rfc2136" />
                            <el-option label="rimuhosting" value="rimuhosting" />
                            <el-option label="route53" value="route53" />
                            <el-option label="safedns" value="safedns" />
                            <el-option label="sakuracloud" value="sakuracloud" />
                            <el-option label="scaleway" value="scaleway" />
                            <el-option label="selectel" value="selectel" />
                            <el-option label="servercow" value="servercow" />
                            <el-option label="shellrent" value="shellrent" />
                            <el-option label="simply" value="simply" />
                            <el-option label="sonic" value="sonic" />
                            <el-option label="stackpath" value="stackpath" />
                            <el-option label="tencentcloud" value="tencentcloud" />
                            <el-option label="transip" value="transip" />
                            <el-option label="ultradns" value="ultradns" />
                            <el-option label="variomedia" value="variomedia" />
                            <el-option label="vegadns" value="vegadns" />
                            <el-option label="vercel" value="vercel" />
                            <el-option label="versio" value="versio" />
                            <el-option label="vinyldns" value="vinyldns" />
                            <el-option label="vkcloud" value="vkcloud" />
                            <el-option label="vscale" value="vscale" />
                            <el-option label="vultr" value="vultr" />
                            <el-option label="webnames" value="webnames" />
                            <el-option label="websupport" value="websupport" />
                            <el-option label="wedos" value="wedos" />
                            <el-option label="yandex" value="yandex" />
                            <el-option label="yandex360" value="yandex360" />
                            <el-option label="yandexcloud" value="yandexcloud" />
                            <el-option label="zoneee" value="zoneee" />
                            <el-option label="zonomi" value="zonomi" />
                        </el-select>

                    </el-form-item>


                    <el-form-item :label="item"
                        v-if="sslSettings.paramsList.length != 0 && sslSettings.type == 0 && sslSettings.challenge == 'dns'"
                        v-for="item in sslSettings.paramsList">
                        <el-input style="width: 240px" :placeholder="item" v-model="dnsApiParams[item]" />

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

        <el-button :element-loading-text="lang.wait_desc" v-loading.fullscreen.lock="fullscreenLoading" id="next" style="margin-top: 12px" @click="next">{{
            lang.next }}</el-button>

    </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import lang from '../i18n/i18n';
import axios from 'axios'

import { getCurrentInstance } from 'vue'
const app = getCurrentInstance()
const $http = app.appContext.config.globalProperties.$http

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
    "provider": "",
    "challenge": "http",
    "key_path": "./config/ssl/private.key",
    "crt_path": "./config/ssl/public.crt",
    "paramsList": {},
})

const dnsApiParams = reactive({})

const active = ref(0)
const fullscreenLoading = ref(false)


const dnsInfos = ref([
])

const port = ref(80)

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
            sslSettings.type = res.data.type
            if (sslSettings.type == "2"){
                sslSettings.type = "0"
                sslSettings.challenge="dns"
            }


            port.value = res.data.port
        }
    })
}


const setSSLConfig = () => {
    fullscreenLoading.value = true;

    let sslType = sslSettings.type;
    if (sslType == "0" && sslSettings.challenge == "dns") {
        sslType = "2"
    }

    if (sslType == "2") {

        let params = { "action": "setParams", "step": "ssl", };

        params = Object.assign(params, dnsApiParams);

        // dns验证方式先提交DNS api Key
        $http.post("/api/setup", params).then((res) => {
            if (res.errorNo != 0) {
                fullscreenLoading.value = false;
                ElMessage.error(res.errorMsg);
                return;
            }
        })


    }



    $http.post("/api/setup", {
        "action": "set",
        "step": "ssl",
        "ssl_type": sslType,
        "key_path": sslSettings.key_path,
        "crt_path": sslSettings.crt_path,
        "serviceName": sslSettings.provider
    }).then((res) => {
        if (res.errorNo != 0) {
            fullscreenLoading.value = false;
            ElMessage.error(res.errorMsg)
        } else {
            checkStatus();

        }
    })
}


const checkStatus = () => {
    axios.post("/api/ping", {}).then((res) => {
        if (res.data.errorNo != 0) {
            setTimeout(function () {
                checkStatus()
            }, 1000);
        } else {
            window.location.href = "https://" + domainSettings.web_domain;
        }
    }).catch((error) => {
        setTimeout(function () {
            checkStatus()
        }, 1000);
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

const provide_change = () => {
    console.log(sslSettings.provider)
    $http.post("/api/setup", { "action": "getParams", "step": "ssl", "serverName": sslSettings.provider }).then((res) => {
        if (res.errorNo != 0) {
            ElMessage.error(res.errorMsg)
        } else {
            sslSettings.paramsList = res.data
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