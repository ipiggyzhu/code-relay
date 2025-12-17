import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  build: {
    // 禁用生产模式下的 minify 可以防止绑定代码被错误移除
    // 同时保持代码可读性，便于调试
    minify: false,
    rollupOptions: {
      // 完全禁用 tree-shaking 以确保 Wails 绑定不被移除
      treeshake: false
    }
  }
})
