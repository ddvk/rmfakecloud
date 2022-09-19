/** @type {import('vite').UserConfig} */

import * as path from 'path'

import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'
import { VitePluginFonts } from 'vite-plugin-fonts'
import { createHtmlPlugin } from 'vite-plugin-html'
import viteImagemin from 'vite-plugin-imagemin'
import { viteStaticCopy } from 'vite-plugin-static-copy'

import config from './config.js'
import postcss from './postcss.config.js'

const { imagemin } = config

// https://vitejs.dev/config/
export default defineConfig({
  server: {
    proxy: {
      '/ui/api': {
        target: 'http://localhost:3000',
        changeOrigin: true
      }
    }
  },
  plugins: [
    react(),
    viteImagemin(imagemin),
    viteStaticCopy({
      targets: [
        {
          src: 'node_modules/pdfjs-dist/build/pdf.worker.min.js',
          dest: 'lib/'
        },
        {
          src: 'node_modules/pdfjs-dist/cmaps/',
          dest: 'lib/'
        }
      ]
    }),
    createHtmlPlugin({
      minify: true,
      entry: '/src/main.tsx'
    }),
    VitePluginFonts({
      // Custom fonts
      custom: {
        families: [
          {
            name: 'CascadiaCodePL',
            src: './src/assets/fonts/*.woff2'
          }
        ],
        display: 'swap',
        preload: true,
        prefetch: false,
        injectTo: 'head-prepend'
      }
    })
  ],
  css: {
    postcss
  },
  resolve: {
    alias: [
      { find: '@/', replacement: '/src' },
      { find: '@/Assets', replacement: '/src/assets' },
      { find: '@/Components', replacement: '/src/components' },
      { find: '@/API', replacement: '/src/api' }
    ]
  },
  optimizeDeps: {
    esbuildOptions: {
      plugins: [
        {
          name: 'resolve-fixup',
          setup(build) {
            build.onResolve({ filter: /react-virtualized/ }, async () => {
              return {
                path: path.resolve('./node_modules/react-virtualized/dist/umd/react-virtualized.js')
              }
            })
          }
        }
      ]
    }
  }
})
