import { join } from "path";
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import postcssPluginPx2rem from "postcss-plugin-px2rem";

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [vue({
        reactivityTransform: true,
    }),],
    resolve: {
        alias: {
            '@': join(__dirname, "src"),
        }
    },
    server: {
        host: '0.0.0.0',
        open: false,
        port: 9200,
        strictPort: true,
        proxy: {
          '/v1': {
            target: 'http://127.0.0.1:9000',
            changeOrigin: true,
          },
        }
    },
    build: {
        outDir: "dist",
        assetsDir: "static",
        assetsInlineLimit: 150000
    },
    css: {
        preprocessorOptions: {
            less: {
                charset: false,
                additionalData: '@import "./src/style/global.less";',
            }
        },
        postcss: {
            plugins: [
                postcssPluginPx2rem({
                    rootValue: 37.5,
                    exclude: /(node_module)/,
                    mediaQuery: false,
                }),
            ]
        },
    }
})
