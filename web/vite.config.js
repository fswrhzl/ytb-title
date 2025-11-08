import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";

// https://vite.dev/config/
export default defineConfig({
    plugins: [vue()],
    server: {
        proxy: {
            "/api": {
                target: "http://localhost:60000",
                changeOrigin: true,
                secure: false,
                // rewrite: (path) => path.replace(/^\/api/, ""),
            },
        },
    },
    build: {
        // 构建输出目录
        outDir: "../web/dist",
    },
});
