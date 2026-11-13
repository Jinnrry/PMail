<template>
  <iframe
    ref="frame"
    class="mail-body-frame"
    title="Email body"
    :srcdoc="srcdoc"
    :style="{ height: `${height}px` }"
    sandbox="allow-same-origin allow-popups allow-popups-to-escape-sandbox"
    referrerpolicy="no-referrer"
    @load="syncHeight"
  />
</template>

<script setup>
import {computed, onBeforeUnmount, onMounted, ref, watch} from 'vue'

const props = defineProps({
  html: {type: String, default: ''}
})

const frame = ref(null)
const height = ref(200)
const colors = ref(readColors())
const colorScheme = window.matchMedia('(prefers-color-scheme: dark)')
let resizeObserver

function readColors() {
  const style = getComputedStyle(document.documentElement)
  return {
    text: style.getPropertyValue('--pm-text-primary').trim() || 'inherit',
    link: style.getPropertyValue('--pm-primary-color').trim() || 'inherit'
  }
}

const srcdoc = computed(() => `<base target="_blank">
<style>
  html, body {
    margin: 0;
    font-size: 16px;
    line-height: 1.7;
    color: ${colors.value.text};
    background: transparent;
    overflow-wrap: anywhere;
  }
  img { max-width: 100%; height: auto; }
  a { color: ${colors.value.link}; }
</style>
${props.html}`)

function disconnect() {
  resizeObserver?.disconnect()
  resizeObserver = null
}

function syncHeight() {
  disconnect()
  const doc = frame.value.contentDocument
  const updateHeight = () => {
    height.value = Math.max(200, doc.documentElement.scrollHeight)
  }
  updateHeight()
  resizeObserver = new ResizeObserver(updateHeight)
  resizeObserver.observe(doc.body)
}

function syncColors() {
  colors.value = readColors()
}

watch(srcdoc, () => {
  disconnect()
  height.value = 200
})

onMounted(() => {
  colorScheme.addEventListener('change', syncColors)
})

onBeforeUnmount(() => {
  disconnect()
  colorScheme.removeEventListener('change', syncColors)
})
</script>

<style scoped>
.mail-body-frame {
  display: block;
  width: 100%;
  min-height: 200px;
  border: 0;
  background: transparent;
}
</style>
