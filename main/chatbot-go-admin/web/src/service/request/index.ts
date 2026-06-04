import type { AxiosResponse } from 'axios';
import { BACKEND_ERROR_CODE, createFlatRequest, createRequest } from '@sa/axios';
import { useAuthStore } from '@/store/modules/auth';
import { localStg } from '@/utils/storage';
import { getServiceBaseURL } from '@/utils/service';
import { $t } from '@/locales';
import { getAuthorization, handleExpiredRequest, showErrorMsg } from './shared';
import type { RequestInstanceState } from './type';

const { baseURL, otherBaseURL } = getServiceBaseURL(import.meta.env);

/**
 * 前后端统一 HTTP 协议（强制）
 *
 * ## 统一响应包体（后端必须遵守）
 * - **HTTP Status**：用于表达“协议/鉴权/资源”的语义（例如 401/403/404/405/500）
 * - **JSON Envelope**：用于表达“业务成功/失败”
 *
 * ```ts
 * type Envelope<T> = { code: number; message: string; data: T | null }
 * ```
 *
 * - **成功**：`code === 0`，`message === "ok"`，`data` 为业务数据（允许为 `null`）
 * - **失败**：`code !== 0`，`message` 为可展示错误信息，`data` 为 `null`
 *
 * ## 前端职责（必须在工具层完成）
 * - 自动注入 `Authorization: Bearer <token>`
 * - 解包 `Envelope<T>`，并把“异常提示/登出/刷新 token”统一在这里处理
 * - 页面 / Store / 组件 **禁止再做** HTTP 异常处理（不要再写 try/catch 去弹错误）
 *
 * 规范文档：
 * - `rules/web/HTTP_CONTRACT.md`
 * - `rules/server/HTTP_CONTRACT.md`
 */

export const requestRaw = createFlatRequest(
  {
    baseURL
  },
  {
    defaultState: {
      errMsgStack: [],
      refreshTokenPromise: null
    } as RequestInstanceState,
    transform(response: AxiosResponse<App.Service.Response<any>>) {
      return response.data.data;
    },
    async onRequest(config) {
      const Authorization = getAuthorization();
      Object.assign(config.headers, { Authorization });

      return config;
    },
    isBackendSuccess(response: AxiosResponse<App.Service.Response<any>>) {
      // when the backend response code is "0000"(default), it means the request is success
      // to change this logic by yourself, you can modify the `VITE_SERVICE_SUCCESS_CODE` in `.env` file
      return String(response.data.code) === import.meta.env.VITE_SERVICE_SUCCESS_CODE;
    },
    async onBackendFail(response: AxiosResponse<App.Service.Response<any>>, instance) {
      const authStore = useAuthStore();
      const responseCode = String(response.data.code);

      function handleLogout() {
        authStore.resetStore();
      }

      function logoutAndCleanup() {
        handleLogout();
        window.removeEventListener('beforeunload', handleLogout);

        requestRaw.state.errMsgStack = requestRaw.state.errMsgStack.filter((msg: string) => msg !== response.data.message);
      }

      // when the backend response code is in `logoutCodes`, it means the user will be logged out and redirected to login page
      const logoutCodes = import.meta.env.VITE_SERVICE_LOGOUT_CODES?.split(',') || [];
      if (logoutCodes.includes(responseCode)) {
        handleLogout();
        return null;
      }

      // when the backend response code is in `modalLogoutCodes`, it means the user will be logged out by displaying a modal
      const modalLogoutCodes = import.meta.env.VITE_SERVICE_MODAL_LOGOUT_CODES?.split(',') || [];
      if (modalLogoutCodes.includes(responseCode) && !requestRaw.state.errMsgStack?.includes(response.data.message)) {
        requestRaw.state.errMsgStack = [...(requestRaw.state.errMsgStack || []), response.data.message];

        // prevent the user from refreshing the page
        window.addEventListener('beforeunload', handleLogout);

        window.$dialog?.error({
          title: $t('common.error'),
          content: response.data.message,
          positiveText: $t('common.confirm'),
          maskClosable: false,
          closeOnEsc: false,
          onPositiveClick() {
            logoutAndCleanup();
          },
          onClose() {
            logoutAndCleanup();
          }
        });

        return null;
      }

      // when the backend response code is in `expiredTokenCodes`, it means the token is expired, and refresh token
      // the api `refreshToken` can not return error code in `expiredTokenCodes`, otherwise it will be a dead loop, should return `logoutCodes` or `modalLogoutCodes`
      const expiredTokenCodes = import.meta.env.VITE_SERVICE_EXPIRED_TOKEN_CODES?.split(',') || [];
      if (expiredTokenCodes.includes(responseCode)) {
        const success = await handleExpiredRequest(requestRaw.state);
        if (success) {
          const Authorization = getAuthorization();
          Object.assign(response.config.headers, { Authorization });

          return instance.request(response.config) as Promise<AxiosResponse>;
        }
      }

      return null;
    },
    onError(error) {
      // when the request is fail, you can show error message

      let message = error.message;
      let backendErrorCode = '';

      // get backend error message and code
      if (error.code === BACKEND_ERROR_CODE) {
        message = error.response?.data?.message || message;
        backendErrorCode = String(error.response?.data?.code || '');
      }

      // the error message is displayed in the modal
      const modalLogoutCodes = import.meta.env.VITE_SERVICE_MODAL_LOGOUT_CODES?.split(',') || [];
      if (modalLogoutCodes.includes(backendErrorCode)) {
        return;
      }

      // when the token is expired, refresh token and retry request, so no need to show error message
      const expiredTokenCodes = import.meta.env.VITE_SERVICE_EXPIRED_TOKEN_CODES?.split(',') || [];
      if (expiredTokenCodes.includes(backendErrorCode)) {
        return;
      }

      showErrorMsg(requestRaw.state, message);
    }
  }
);

type RequestFallback<T> = T | (() => T);

/**
 * 业务请求入口（推荐）。
 *
 * - 永远返回 **已解包后的业务数据**（即 `Envelope.data`）
 * - 若失败：工具层已完成提示/登出/重试等逻辑，这里返回 fallback（避免页面出现 `undefined.map`）
 */
export async function request<T>(config: Parameters<typeof requestRaw<T>>[0], fallback: RequestFallback<T>): Promise<T> {
  const res = await requestRaw<T>(config);
  if (!res.error && res.data !== undefined) {
    return res.data as T;
  }
  return typeof fallback === 'function' ? (fallback as () => T)() : fallback;
}

export const demoRequest = createRequest(
  {
    baseURL: otherBaseURL.demo
  },
  {
    transform(response: AxiosResponse<App.Service.DemoResponse>) {
      return response.data.result;
    },
    async onRequest(config) {
      const { headers } = config;

      // set token
      const token = localStg.get('token');
      const Authorization = token ? `Bearer ${token}` : null;
      Object.assign(headers, { Authorization });

      return config;
    },
    isBackendSuccess(response) {
      // when the backend response code is "200", it means the request is success
      // you can change this logic by yourself
      return response.data.status === '200';
    },
    async onBackendFail(_response) {
      // when the backend response code is not "200", it means the request is fail
      // for example: the token is expired, refresh token and retry request
    },
    onError(error) {
      // when the request is fail, you can show error message

      let message = error.message;

      // show backend error message
      if (error.code === BACKEND_ERROR_CODE) {
        message = error.response?.data?.message || message;
      }

      window.$message?.error(message);
    }
  }
);
