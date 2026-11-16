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
import {computed, onBeforeUnmount, ref, watch} from 'vue'

const props = defineProps({
  html: {type: String, default: ''}
})

const frame = ref(null)
const maxHeight = 1200
const height = ref(200)
let resizeObserver

const srcdoc = computed(() => `<!DOCTYPE html>
<base target="_blank">
<style>
  html, body {
    margin: 0;
    font-size: 16px;
    line-height: 1.7;
    color: #1d1d1f;
    background: transparent;
    overflow-wrap: anywhere;
  }
  img { max-width: 100%; height: auto; }
  a { color: #0071e3; }
  @media (prefers-color-scheme: dark) {
    html, body { color: #f5f7fa; }
  }
</style>
${props.html}`)

function disconnect() {
  resizeObserver?.disconnect()
  resizeObserver = null
}

function syncHeight() {
  disconnect()
  const doc = frame.value?.contentDocument
  // A message link can navigate the frame to a different origin.
  if (!doc?.documentElement || !doc.body) return
  const updateHeight = () => {
    if (frame.value?.contentDocument !== doc) return
    // Viewport-relative email CSS can grow with the frame on every resize.
    // Bound auto-sizing and let longer content scroll inside the iframe.
    height.value = Math.min(maxHeight, Math.max(200, doc.documentElement.scrollHeight))
  }
  updateHeight()
  resizeObserver = new ResizeObserver(updateHeight)
  resizeObserver.observe(doc.body)
}

watch(srcdoc, () => {
  disconnect()
  height.value = 200
})

onBeforeUnmount(disconnect)
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
