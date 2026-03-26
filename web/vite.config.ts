import path from 'node:path';

import { TDesignResolver } from '@tdesign-vue-next/auto-import-resolver';
import vue from '@vitejs/plugin-vue';
import vueJsx from '@vitejs/plugin-vue-jsx';
import AutoImport from 'unplugin-auto-import/vite';
import Components from 'unplugin-vue-components/vite';
import type { ConfigEnv, UserConfig } from 'vite';
import { loadEnv } from 'vite';
import { viteMockServe } from 'vite-plugin-mock';
import svgLoader from 'vite-svg-loader';

const CWD = process.cwd();

// https://vitejs.dev/config/
export default ({ mode }: ConfigEnv): UserConfig => {
  const { VITE_BASE_URL, VITE_API_URL_PREFIX, VITE_API_URL } = loadEnv(mode, CWD);
  return {
    base: VITE_BASE_URL,
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },

    css: {
      preprocessorOptions: {
        less: {
          modifyVars: {
            hack: `true; @import (reference) "${path.resolve('src/style/variables.less')}";`,
          },
          math: 'strict',
          javascriptEnabled: true,
        },
      },
    },

    plugins: [
      vue(),
      vueJsx(),
      AutoImport({
        dts: 'src/auto-imports.d.ts',
        resolvers: [TDesignResolver({ library: 'vue-next' })],
      }),
      Components({
        dts: 'src/components.d.ts',
        resolvers: [TDesignResolver({ library: 'vue-next' })],
      }),
      viteMockServe({
        mockPath: 'mock',
        enable: mode === 'mock',
      }),
      svgLoader(),
    ],

    build: {
      emptyOutDir: true,
      sourcemap: false,
      rollupOptions: {
        output: {
          manualChunks(id) {
            if (!id.includes('node_modules'))
              return;

            if (id.includes('echarts'))
              return 'vendor-echarts';
            if (id.includes('tdesign-vue-next') || id.includes('tdesign-icons-vue-next'))
              return 'vendor-tdesign';
            if (id.includes('vue-router') || id.includes('pinia') || id.includes('vue-i18n') || id.includes('/vue/'))
              return 'vendor-vue';
          },
        },
      },
    },

    server: {
      port: 3002,
      host: '0.0.0.0',
      proxy: {
        [VITE_API_URL_PREFIX]: {
          target: VITE_API_URL || 'http://127.0.0.1:8080',
          changeOrigin: true,
        },
      },
    },
  };
};
