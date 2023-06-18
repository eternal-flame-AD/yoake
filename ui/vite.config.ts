import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  base: "/ui/",
  plugins: [react()],
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:3000",
      }
    }
  }
})
