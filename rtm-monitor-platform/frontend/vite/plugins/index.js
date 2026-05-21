import vue from '@vitejs/plugin-vue'
import AutoImport from 'unplugin-auto-import/vite'
import VueSetupExtend from 'unplugin-vue-setup-extend-plus/vite'
import { createSvgIconsPlugin } from 'vite-plugin-svg-icons'
import compression from 'vite-plugin-compression'
import path from 'path'

export default function createVitePlugins(env, isBuild) {
  const vitePlugins = [
    // vue支持
    vue(),
    // setup语法糖组件名支持
    VueSetupExtend(),
    // 自动导入API
    AutoImport({
      imports: ['vue', 'vue-router', 'pinia'],
      dts: false
    }),
    // SVG图标
    createSvgIconsPlugin({
      iconDirs: [path.resolve(process.cwd(), 'src/assets/icons/svg')],
      symbolId: 'icon-[name]'
    })
  ]

  // 生产环境开启gzip压缩
  if (isBuild) {
    vitePlugins.push(
      compression({
        verbose: true,
        disable: false,
        threshold: 10240,
        algorithm: 'gzip',
        ext: '.gz'
      })
    )
  }

  return vitePlugins
}
