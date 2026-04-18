<template>
  <div class="composer-container">
    <div class="composer-header">
      <h2 class="composer-title">{{ lang.compose }}</h2>
      <div class="composer-actions">
        <el-button class="action-btn" @click="upload">
          <el-icon><Paperclip /></el-icon> {{ lang.add_att }}
        </el-button>
        <el-button type="primary" class="send-btn" @click="send(ruleFormRef)">
          <el-icon><Position /></el-icon> {{ lang.send }}
        </el-button>
        <input v-show="false" ref="fileRef" type="file" @change="fileChange" multiple>
      </div>
    </div>

    <div class="composer-body">
      <el-form :rules="rules" ref="ruleFormRef" :model="ruleForm" class="composer-form">
        
        <div class="compose-row">
          <div class="field-label">{{ lang.sender }}</div>
          <el-form-item prop="sender" class="flex-grow-item m-0">
            <el-popover trigger="click" :width="400" placement="bottom-start">
              <template #reference>
                <div class="sender-selector">
                  <span class="sender-name">{{ ruleForm.nickName }}</span>
                  <span class="sender-email">&lt;{{ ruleForm.sender }}@{{ ruleForm.pickDomain }}&gt;</span>
                  <el-icon class="arrow-down"><ArrowDown /></el-icon>
                </div>
              </template>
              <template #default>
                <div class="sender-edit-card">
                  <div class="edit-row">
                    <span class="edit-label">Prefix</span>
                    <el-input 
                      :disabled="!(globalStatus.userInfos.is_admin)"
                      v-model="ruleForm.sender" 
                      :placeholder="lang.sender_desc"
                    />
                  </div>
                  <div class="edit-row">
                    <span class="edit-label">Domain</span>
                    <el-select v-model="ruleForm.pickDomain" class="w-full">
                      <el-option :value="item" v-for="item in ruleForm.domains" :key="item">{{ item }}</el-option>
                    </el-select>
                  </div>
                  <div class="edit-row">
                    <span class="edit-label">{{ lang.nick_name }}</span>
                    <el-input v-model="ruleForm.nickName"/>
                  </div>
                </div>
              </template>
            </el-popover>
          </el-form-item>
        </div>

        <div class="compose-row">
          <div class="field-label">{{ lang.to }}</div>
          <el-form-item prop="receivers" class="flex-grow-item m-0 border-less-select">
            <el-select 
              v-model="ruleForm.receivers" 
              multiple filterable allow-create :reserve-keyword="false" 
              placeholder="Recipients..."
            ></el-select>
          </el-form-item>
        </div>

        <div class="compose-row" v-if="ruleForm.cc.length > 0 || showCcBcc">
          <div class="field-label">{{ lang.cc }}</div>
          <el-form-item prop="cc" class="flex-grow-item m-0 border-less-select">
            <el-select 
              v-model="ruleForm.cc" 
              multiple filterable allow-create :reserve-keyword="false" 
              placeholder="Cc..."
            ></el-select>
          </el-form-item>
        </div>

        <div class="compose-row" v-if="ruleForm.bcc.length > 0 || showCcBcc">
          <div class="field-label">{{ lang.bcc }}</div>
          <el-form-item prop="bcc" class="flex-grow-item m-0 border-less-select">
            <el-select 
              v-model="ruleForm.bcc" 
              multiple filterable allow-create :reserve-keyword="false" 
              placeholder="Bcc..."
            ></el-select>
          </el-form-item>
        </div>

        <div class="compose-row subject-row">
          <el-form-item prop="subject" class="w-full m-0 border-less-input">
            <el-input v-model="ruleForm.subject" placeholder="Subject"></el-input>
            <div class="cc-bcc-toggle" @click="showCcBcc = !showCcBcc" v-if="!showCcBcc">Cc/Bcc</div>
          </el-form-item>
        </div>

        <div class="editor-wrapper">
          <Toolbar class="editor-toolbar" :editor="editorRef" :defaultConfig="toolbarConfig" :mode="mode"/>
          <Editor class="editor-content" v-model="valueHtml" :defaultConfig="editorConfig" :mode="mode" @onCreated="handleCreated"/>
        </div>

        <div class="attachments-preview" v-if="fileList.length > 0">
          <div class="att-chip" v-for="(item, index) in fileList" :key="index">
            <el-icon class="att-icon"><Document /></el-icon>
            <span class="att-name">{{ item.name }}</span>
            <el-icon class="att-remove" @click="delFile(index)"><Close/></el-icon>
          </div>
        </div>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import '@wangeditor/editor/dist/css/style.css'
import {ElMessage} from 'element-plus'
import {onBeforeUnmount, reactive, ref, shallowRef} from 'vue'
import {Close, Paperclip, Position, ArrowDown, Document} from '@element-plus/icons-vue';
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
const showCcBcc = ref(false);

if (lang.lang === "zhCn") {
  i18nChangeLanguage('zh-CN')
} else {
  i18nChangeLanguage('en')
}

// 内容 HTML
const valueHtml = ref('<p><br></p>')

const toolbarConfig = {}
const editorConfig = {
  MENU_CONF: {},
  placeholder: 'Write your message...'
}

editorConfig.MENU_CONF['uploadImage'] = {
  base64LimitSize: 100 * 1024 * 1024 * 1024,
}
const mode = ref()
const fileRef = ref();
const ruleFormRef = ref()
const ruleForm = reactive({
  nickName: '',
  sender: '',
  receivers: [],
  cc: [],
  bcc: [],
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
    if (!checkEmail(element)) {
      callback(new Error(lang.err_email_format))
      return
    }
  }
  callback()
}

const rules = reactive({
  sender: [{validator: validateSender, trigger: 'change'}],
  receivers: [{validator: validateReceivers, trigger: 'change'}],
  cc: [{validator: validateCc, trigger: 'change'}],
  bcc: [{validator: validateBcc, trigger: 'change'}],
  subject: [{required: true, message: lang.err_title_must, trigger: 'change'}],
})

const editorRef = shallowRef()
onBeforeUnmount(() => {
  const editor = editorRef.value
  if (editor == null) return
  editor.destroy()
})

const handleCreated = (editor) => {
  editorRef.value = editor
}

const send = function (formEl) {
  if (!formEl) return
  formEl.validate((valid) => {
    if (valid) {
      if(ruleForm.receivers.length === 0 && ruleForm.cc.length === 0 && ruleForm.bcc.length === 0) {
        ElMessage.warning("Please specify at least one recipient");
        return;
      }
      let objectTos = ruleForm.receivers.map(e => ({name: "", email: e}));
      let objectCcs = ruleForm.cc.map(e => ({name: "", email: e}));
      let objectBccs = ruleForm.bcc.map(e => ({name: "", email: e}));

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
          ElMessage.success(lang.succ_send)
          groupStore.name = lang.outbox
          groupStore.tag = '{"type":1,"status":-1}'
          router.replace({name: 'list'})
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
  if (!files.length) return;
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

<style scoped>
.composer-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--pm-surface-glass);
  border-radius: var(--pm-radius-xl);
  box-shadow: var(--pm-shadow-md);
  border: 1px solid var(--pm-glass-border);
  backdrop-filter: blur(16px);
  overflow: hidden;
  animation: pm-rise-in 0.34s var(--pm-ease-out);
}

.composer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--pm-border-color);
  background: var(--pm-surface-glass-soft);
}

.composer-title {
  font-size: 28px;
  font-weight: 600;
  color: var(--pm-text-primary);
  margin: 0;
  letter-spacing: -0.02em;
}

.composer-actions {
  display: flex;
  gap: 12px;
}

.action-btn {
  background: var(--pm-surface-glass);
  border: 1px solid var(--pm-border-color);
  color: var(--pm-text-secondary);
}

.action-btn:hover {
  background: var(--pm-surface-solid);
  color: var(--pm-text-primary);
  border-color: var(--pm-border-strong);
}

.send-btn {
  box-shadow: 0 8px 16px rgba(0, 113, 227, 0.2);
}

.composer-body {
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.composer-form {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.compose-row {
  display: flex;
  align-items: center;
  border-bottom: 1px solid var(--pm-border-subtle);
  padding: 6px 20px;
}

.field-label {
  width: 70px;
  color: var(--pm-text-secondary);
  font-weight: 600;
  font-size: 13px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.flex-grow-item {
  flex-grow: 1;
}

.m-0 { margin: 0; }
.w-full { width: 100%; }

.sender-selector {
  display: inline-flex;
  align-items: center;
  padding: 8px 12px;
  border-radius: 12px;
  cursor: pointer;
  background-color: var(--pm-surface-muted);
  border: 1px solid var(--pm-border-color);
  transition: background 0.2s;
}

.sender-selector:hover {
  background-color: var(--pm-surface-solid);
  transform: translateY(-1px);
}

.sender-name {
  font-weight: 600;
  margin-right: 6px;
  color: var(--pm-text-primary);
}

.sender-email {
  color: var(--pm-text-secondary);
  margin-right: 8px;
}

.arrow-down {
  font-size: 12px;
  color: var(--pm-text-muted);
}

.sender-edit-card {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.edit-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.edit-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--pm-text-secondary);
}

/* Borderless inputs to make it look like integrated fields */
.border-less-select :deep(.el-input__wrapper),
.border-less-input :deep(.el-input__wrapper) {
  box-shadow: none !important;
  background: transparent;
  padding-left: 0;
}

.border-less-select :deep(.el-select .el-input__wrapper) {
  background: var(--pm-surface-muted) !important;
  box-shadow: 0 0 0 1px var(--pm-border-color) inset !important;
  border-radius: 10px;
  padding: 2px 10px;
}

.border-less-select :deep(.el-select .el-input__wrapper:hover),
.border-less-select :deep(.el-select .el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--pm-primary-color) inset !important;
  background: var(--pm-surface-solid) !important;
}

.border-less-select :deep(.el-select__tags .el-tag) {
  background: var(--pm-surface-glass);
  border-color: var(--pm-border-color);
  color: var(--pm-text-primary);
}

.border-less-select :deep(.el-input__inner),
.border-less-input :deep(.el-input__inner) {
  font-size: 15px;
}

.subject-row {
  position: relative;
}

.cc-bcc-toggle {
  position: absolute;
  right: 0;
  top: 50%;
  transform: translateY(-50%);
  font-size: 13px;
  color: var(--pm-primary-color);
  cursor: pointer;
  font-weight: 500;
}

.cc-bcc-toggle:hover {
  text-decoration: underline;
}

.editor-wrapper {
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  min-height: 300px;
  --w-e-toolbar-bg-color: var(--pm-surface-glass-soft);
  --w-e-toolbar-color: var(--pm-text-primary);
  --w-e-toolbar-active-color: var(--pm-primary-color);
  --w-e-toolbar-active-bg-color: var(--pm-row-hover);
  --w-e-textarea-bg-color: transparent;
  --w-e-textarea-color: var(--pm-text-primary);
  --w-e-textarea-slight-bg-color: transparent;
  --w-e-textarea-slight-color: var(--pm-text-muted);
  --w-e-border-color: var(--pm-border-subtle);
}

.editor-toolbar {
  border-bottom: 1px solid var(--pm-border-subtle);
  background-color: var(--pm-surface-glass-soft) !important;
}

.editor-content {
  flex-grow: 1;
  padding: 14px 20px;
  overflow-y: hidden;
}

.editor-content :deep(.w-e-text-container),
.editor-content :deep(.w-e-scroll),
.editor-content :deep(.w-e-text) {
  background: transparent !important;
  color: var(--pm-text-primary) !important;
}

.editor-content :deep(.w-e-text-placeholder) {
  color: var(--pm-text-muted) !important;
}

.attachments-preview {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 14px 20px;
  border-top: 1px solid var(--pm-border-color);
  background: var(--pm-surface-glass-soft);
}

.att-chip {
  display: flex;
  align-items: center;
  background: var(--pm-bg-secondary);
  border: 1px solid var(--pm-border-color);
  padding: 6px 12px;
  border-radius: 999px;
  font-size: 13px;
  box-shadow: var(--pm-shadow-sm);
  transition: transform 0.2s var(--pm-ease-out), box-shadow 0.2s var(--pm-ease-out);
}

.att-chip:hover {
  transform: translateY(-1px);
  box-shadow: var(--pm-shadow-md);
}

.att-icon {
  margin-right: 6px;
  color: var(--pm-text-secondary);
}

.att-name {
  max-width: 150px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-right: 8px;
}

.att-remove {
  cursor: pointer;
  color: var(--pm-text-muted);
}

.att-remove:hover {
  color: #ef4444;
}

@media (prefers-color-scheme: dark) {
  .editor-wrapper {
    background: var(--pm-surface-glass-soft);
  }

  .editor-content :deep(.w-e-text-container) {
    border-left-color: transparent !important;
    border-right-color: transparent !important;
  }

  .editor-content :deep(.w-e-bar) {
    background: var(--pm-surface-glass-soft) !important;
    color: var(--pm-text-primary) !important;
  }

  .editor-content :deep(.w-e-bar-item button),
  .editor-content :deep(.w-e-menu-tooltip-v5) {
    color: var(--pm-text-primary) !important;
  }
}

@media (max-width: 768px) {
  .sender-selector {
    padding: 6px;
    max-width: 100%;
    overflow: hidden;
  }
  .sender-email {
    display: none;
  }
  .composer-header {
    padding: 12px 16px;
  }
}
</style>