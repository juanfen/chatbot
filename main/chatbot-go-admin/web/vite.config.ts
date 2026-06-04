import process from 'node:process';
import { URL, fileURLToPath } from 'node:url';
import type { ProxyOptions } from 'vite';
import { defineConfig, loadEnv } from 'vite';
import { setupVitePlugins } from './build/plugins';
import { getBuildTime } from './build/config';

/** 开发态将 /auth、/route、/api 等转发到远程 admin server，避免跨域 */
function buildDevProxy(env: Env.ImportMeta, isServe: boolean): Record<string, ProxyOptions> | undefined {
  if (!isServe) return undefined;
  const target = env.VITE_DEV_PROXY_TARGET?.trim();
  if (!target) return undefined;
  const common: ProxyOptions = { target, changeOrigin: true };
  return {
    '/auth': common,
    '/route': common,
    '/api': common,
    '/sys': common,
    '/biz': common,
    '/mobile': common,
    '/device': common
  };
}

export default defineConfig(configEnv => {
  const viteEnv = loadEnv(configEnv.mode, process.cwd()) as unknown as Env.ImportMeta;

  const buildTime = getBuildTime();

  return {
    base: viteEnv.VITE_BASE_URL,
    resolve: {
      alias: {
        '~': fileURLToPath(new URL('./', import.meta.url)),
        '@': fileURLToPath(new URL('./src', import.meta.url))
      }
    },
    css: {
      preprocessorOptions: {
        scss: {
          api: 'modern-compiler',
          additionalData: `@use "@/styles/scss/global.scss" as *;`
        }
      }
    },
    plugins: setupVitePlugins(viteEnv, buildTime),
    define: {
      BUILD_TIME: JSON.stringify(buildTime)
    },
    server: {
      host: '0.0.0.0',
      port: 19527,
      open: true,
      // 本机 dev 时 VITE_SERVICE_BASE_URL 留空，请求走同源 + 代理，避免浏览器 CORS
      proxy: buildDevProxy(viteEnv, configEnv.command === 'serve')
    },
    preview: {
      port: 9725
    },
    build: {
      reportCompressedSize: false,
      sourcemap: viteEnv.VITE_SOURCE_MAP === 'Y',
      commonjsOptions: {
        ignoreTryCatch: false
      }
    }
  };
});
