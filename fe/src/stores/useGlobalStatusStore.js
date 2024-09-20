import {defineStore} from "pinia";

const useGlobalStatusStore = defineStore('useGlobalStatusStore', {
    state() {
        return {
            isLogin: true,
            userInfos:{}
        }
    },
    getters: {},
    actions: {}
})


export {useGlobalStatusStore};