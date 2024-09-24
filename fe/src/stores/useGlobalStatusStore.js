import {defineStore} from "pinia";
import {http} from "@/utils/axios";

const useGlobalStatusStore = defineStore('useGlobalStatusStore', {
    state() {
        return {
            userInfos: {}
        }
    },
    getters: {
        isLogin(state) {
            return Object.keys(state.userInfos) !== 0
        }
    },
    actions: {
        init(callback) {
            let that = this
            http.post("/api/user/info", {}).then(res => {
                if (res.errorNo === 0) {
                    Object.assign(that.userInfos, res.data)
                    console.log("userInfos")
                    callback()
                }
            })
        }
    }
})


export {useGlobalStatusStore};