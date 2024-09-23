import { ref } from 'vue'
import { defineStore } from 'pinia'
import lang from '../i18n/i18n';

const useGroupStore = defineStore('group', () => {
  const tag = ref("")
  const name = ref(lang.inbox)
  return { tag, name }
})

export default useGroupStore