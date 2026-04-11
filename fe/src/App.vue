<script setup>
import {RouterView, useRoute} from 'vue-router'
import HomeHeader from '@/components/HomeHeader.vue'
import HomeAside from '@/components/HomeAside.vue';
import lang from '@/i18n/i18n';
import {ref, watch} from 'vue'

const route = useRoute()
const pageName = ref(route.name)


const showAside = ref(true)


watch(
    () => route.fullPath,
    () => {
      pageName.value = route.name
    }
)

const toggleAside = () => {
  showAside.value = !showAside.value
}
</script>

<template>
  <div id="main">
    <HomeHeader/>
    <button @click="toggleAside" class="el-button el-button--small" v-if="pageName !== 'login' && pageName !== 'setup'">
      {{ showAside ? lang.collapse : lang.expand }}
    </button>
    <div id="content">
      <div id="aside" v-if="pageName !== 'login' && pageName !== 'setup' && showAside">
        <HomeAside/>
      </div>
      <div id="body">
        <RouterView/>
      </div>
    </div>
  </div>
</template>


<style scoped>
#aside {
  background-color: #F1F1F1;
}

#body {
  width: 100%;
  height: 100%;
  flex: 1;
  overflow-y: auto;
}

#content {
  display: flex;
  height: 100%;
}

#main {
  height: 100%;
  display: flex;
  flex-direction: column;
}
</style>
