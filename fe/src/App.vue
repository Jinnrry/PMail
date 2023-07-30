<script setup>
import { RouterView } from 'vue-router'
import HomeHeader from '@/components/HomeHeader.vue'
import HomeAside from '@/components/HomeAside.vue';
import { watch,ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
const route = useRoute()

const pageName = ref(route.name) 

watch(
  () => route.fullPath,
  (n, o) => {
    pageName.value = route.name
  }
)

</script>

<template>
  <div id="main">
    <HomeHeader />
    <div id="content">
      <div id="aside" v-if="pageName != 'login'">
        <HomeAside />
      </div>
      <div id="body">
        <RouterView />
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
