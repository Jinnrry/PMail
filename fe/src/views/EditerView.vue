<template>
  <div id="main">
    <el-form label-width="100px" :rules="rules" ref="ruleFormRef" :model="ruleForm" status-icon>

      <el-form-item :label="lang.sender" prop="sender">
        <el-popover trigger="click" :width="600">
          <template #reference>
            <div
                style="border: 1px solid #dcdfe6; border-radius:3px;height: 30px; line-height: 30px; padding: 0 5px 0 5px;">
              <span style="font-size: 16px; font-weight: bolder;">{{ ruleForm.nickName }}</span>
              <span> &lt;{{ ruleForm.sender }}@{{ ruleForm.pickDomain }}&gt;</span>
            </div>
          </template>
          <template #default>
            <div style="display: flex; flex-direction:column;">
              <div style=" margin-bottom: 10px;">
                <el-form-item :label="lang.sender" prop="sender">
                  <el-input style="max-width: 200px" :disabled="!(globalStatus.userInfos.is_admin)"
                            v-model="ruleForm.sender" :placeholder="lang.sender_desc"/>
                  <div>@</div>
                  <el-select v-model="ruleForm.pickDomain">
                    <el-option :value="item" v-for="item in ruleForm.domains" :key="item">{{ item }}</el-option>
                  </el-select>
                </el-form-item>
              </div>

              <div>
                <el-form-item :label="lang.nick_name">
                  <el-input style="max-width: 300px" v-model="ruleForm.nickName"/>
                </el-form-item>
              </div>

            </div>
          </template>
        </el-popover>

      </el-form-item>


      <el-form-item :label="lang.to" prop="receivers">
        <el-select v-model="ruleForm.receivers" style="width: 100%;" multiple filterable allow-create
                   :reserve-keyword="false" :placeholder="lang.to_desc"></el-select>
      </el-form-item>


      <el-form-item :label="lang.cc" prop="cc">
        <el-select v-model="ruleForm.cc" style="width: 100%;" multiple filterable allow-create
                   :reserve-keyword="false" :placeholder="lang.cc_desc"></el-select>
      </el-form-item>

      <el-form-item :label="lang.bcc" prop="bcc">
        <el-select v-model="ruleForm.bcc" style="width: 100%;" multiple filterable allow-create
                   :reserve-keyword="false" :placeholder="lang.bcc_desc"></el-select>
      </el-form-item>

      <el-form-item :label="lang.title" prop="subject">
        <el-input v-model="ruleForm.subject" :placeholder="lang.title"></el-input>
      </el-form-item>


      <div id="editor">
        <div style="border: 1px solid #ccc">
          <Toolbar style="border-bottom: 1px solid #ccc" :editor="editorRef" :defaultConfig="toolbarConfig"
                   :mode="mode"/>
          <Editor style="height: 300px;" v-model="valueHtml" :defaultConfig="editorConfig" :mode="mode"
                  @onCreated="handleCreated"/>
        </div>
      </div>

      <div id="fileList">
        <ol>
          <li v-for="(item, index) in fileList" :key="item">{{ item.name }}
            <el-icon @click="delFile(index)">
              <Close/>
            </el-icon>
          </li>
        </ol>
      </div>

      <div id="sendButton">
        <el-button type="primary" @click="send(ruleFormRef)">{{ lang.send }}</el-button>
        <!-- <el-button>定时发送</el-button> -->

        <div style="margin-left: 15px">
          <el-button @click="upload">{{ lang.add_att }}</el-button>
          <input v-show="false" ref="fileRef" type="file" @change="fileChange">
        </div>
      </div>

    </el-form>

  </div>
</template>

<style scoped>
#main {
  text-align: left;
  padding-right: 20px;
}

#editor {
  padding-left: 25px;
}

#sendButton {
  padding-left: 25px;
  padding-top: 5px;
  display: flex;
}
</style>


<script setup>
import '@wangeditor/editor/dist/css/style.css' // 引入 css
import {ElMessage} from 'element-plus'
import {onBeforeUnmount, reactive, ref, shallowRef} from 'vue'
import {Close} from '@element-plus/icons-vue';
import lang from '../i18n/i18n';
import {Editor, Toolbar} from '@wangeditor/editor-for-vue'
import {i18nChangeLanguage} from '@wangeditor/editor'
import {useRouter} from 'vue-router';
import {http} from "@/utils/axios";
import useGroupStore from '../stores/group'
import {useGlobalStatusStore} from "@/stores/useGlobalStatusStore";

const router = useRouter();

const groupStore = useGroupStore()

const globalStatus = useGlobalStatusStore();

if (lang.lang === "zhCn") {
  i18nChangeLanguage('zh-CN')
} else {
  i18nChangeLanguage('en')
}


// 内容 HTML
const valueHtml = ref('<p>hello</p>')

const toolbarConfig = {}
const editorConfig = {
  MENU_CONF: {},
  placeholder: ''
}


editorConfig.MENU_CONF['uploadImage'] = {
  base64LimitSize: 100 * 1024 * 1024 * 1024,  // 100G以下的文件都base64传
}
const mode = ref()
const fileRef = ref();
const ruleFormRef = ref()
const ruleForm = reactive({
  nickName: '',
  sender: '',
  receivers: '',
  cc: '',
  bcc: '',
  subject: '',
  domains: [],
  pickDomain: ""
})
const fileList = reactive([]);


const init = function () {
    if ( Object.keys(globalStatus.userInfos)==0 || globalStatus.userInfos === null || globalStatus.userInfos == undefined ){
      globalStatus.init(()=>{
        ruleForm.sender = globalStatus.userInfos.account
        ruleForm.domains = globalStatus.userInfos.domains
        ruleForm.pickDomain = globalStatus.userInfos.domains[0]
        ruleForm.nickName = globalStatus.userInfos.name
      })
    }else{
      console.log(globalStatus.userInfos)
      ruleForm.sender = globalStatus.userInfos.account
      ruleForm.domains = globalStatus.userInfos.domains
      ruleForm.pickDomain = globalStatus.userInfos.domains[0]
      ruleForm.nickName = globalStatus.userInfos.name
    }
}
init()


const validateSender = function (rule, value, callback) {
  if (typeof ruleForm.sender === "undefined" || ruleForm.sender === null || ruleForm.sender.trim() === "") {
    callback(new Error(lang.err_sender_must))
  } else if (ruleForm.sender.includes("@")) {
    callback(new Error(lang.only_prefix))
  } else {
    callback()
  }
}

const checkEmail = function (str) {
  const re = /.+@.+\..+/;
  return re.test(str);
}

const validateReceivers = function (rule, value, callback) {
  for (let index = 0; index < ruleForm.receivers.length; index++) {
    let element = ruleForm.receivers[index];
    if (!checkEmail(element)) {
      callback(new Error(lang.err_email_format))
      return
    }
  }
  callback()
}

const validateCc = function (rule, value, callback) {
  for (let index = 0; index < ruleForm.cc.length; index++) {
    let element = ruleForm.cc[index];
    if (!checkEmail(element)) {
      callback(new Error(lang.err_email_format))
      return
    }
  }
  callback()
}

const validateBcc = function (rule, value, callback) {
  for (let index = 0; index < ruleForm.bcc.length; index++) {
    let element = ruleForm.bcc[index];

    console.log(element)
    if (!checkEmail(element)) {
      callback(new Error(lang.err_email_format))
      return
    }
  }
  callback()
}

const rules = reactive({
  sender: [
    {validator: validateSender, trigger: 'change'}
  ],
  receivers: [
    {validator: validateReceivers, trigger: 'change'}
  ],
  cc: [
    {validator: validateCc, trigger: 'change'}
  ],
  bcc: [
    {validator: validateBcc, trigger: 'change'}
  ],
  subject: [
    {required: true, message: lang.err_title_must, trigger: 'change'},
  ],
})


// 编辑器实例，必须用 shallowRef
const editorRef = shallowRef()
// 组件销毁时，也及时销毁编辑器
onBeforeUnmount(() => {
  const editor = editorRef.value
  if (editor == null) return
  editor.destroy()
})

const handleCreated = (editor) => {
  editorRef.value = editor // 记录 editor 实例，重要！
}

const send = function (formEl) {
  if (!formEl) return
  formEl.validate((valid) => {
    if (valid) {
      let objectTos = []
      for (let index = 0; index < ruleForm.receivers.length; index++) {
        let element = ruleForm.receivers[index];
        objectTos.push({
          name: "",
          email: element
        })
      }

      let objectCcs = []
      for (let index = 0; index < ruleForm.cc.length; index++) {
        let element = ruleForm.cc[index];
        objectCcs.push({
          name: "",
          email: element
        })
      }

      let objectBccs = []
      for (let index = 0; index < ruleForm.bcc.length; index++) {
        let element = ruleForm.bcc[index];
        objectBccs.push({
          name: "",
          email: element
        })
      }

      let text = editorRef.value.getText()

      http.post("/api/email/send", {
        from: {name: ruleForm.nickName, email: ruleForm.sender + "@" + ruleForm.pickDomain},
        to: objectTos,
        cc: objectCcs,
        bcc: objectBccs,
        subject: ruleForm.subject,
        text: text,
        html: valueHtml.value,
        attrs: fileList
      }).then(res => {
        if (res.errorNo === 0) {
          ElMessage({
            message: lang.succ_send,
            type: 'success',
          })
          groupStore.name = lang.outbox
          groupStore.tag = '{"type":1,"status":-1}'
          router.replace({
            name: 'list',
          })
        } else {
          ElMessage.error(res.data)
        }
      })
    } else {
      return false
    }
  })


}


const upload = function () {
  fileRef.value.dispatchEvent(new MouseEvent('click'))
}

const fileChange = function (e) {
  let files = e.target.files || e.dataTransfer.files;
  if (!files.length)
    return;
  for (let i = 0; i < files.length; i++) {
    const reader = new FileReader();
    reader.onload = function fileReadCompleted() {
      fileList.push({
        name: files[i].name,
        data: this.result
      })
    };
    reader.readAsDataURL(files[i]);

  }
}

const delFile = function (index) {
  fileList.splice(index, 1);
}
</script>