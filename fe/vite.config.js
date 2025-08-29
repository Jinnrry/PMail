import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import vue from '@vitejs/plugin-vue'
import fs from 'fs'
import path from 'path'

// Function to read and parse config.json
const readConfig = () => {
  const __dirname = path.dirname(fileURLToPath(import.meta.url));
  const configPath = path.resolve(__dirname, '../server/config/config.json');
  if (fs.existsSync(configPath)) {
    const configFile = fs.readFileSync(configPath, 'utf-8');
    try {
      return JSON.parse(configFile);
    } catch (e) {
      console.error('Error parsing config.json:', e);
      return {};
    }
  }
  return {};
};

const config = readConfig();
const frontendPort = config.frontendPort || 5173;
const httpPort = config.httpPort || 80;

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    AutoImport({
      resolvers: [ElementPlusResolver()],
    }),
    Components({
      resolvers: [ElementPlusResolver()],
    }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    port: frontendPort,
    cors: true,
    proxy: {
      "/api": `http://127.0.0.1:${httpPort}/`,
      "/attachments":`http://127.0.0.1:${httpPort}/`
    }
  }
})
