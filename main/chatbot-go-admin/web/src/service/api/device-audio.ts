import axios from 'axios';
import { getServiceBaseURL } from '@/utils/service';
import { getAuthorization } from '../request/shared';
import { request, requestRaw } from '../request';

/** 开发态 baseURL 为空时用 VITE_DEV_PROXY_TARGET 直连后端，避免误以为只请求了 Vite 端口。 */
function resolveApiBase(): string {
  const { baseURL } = getServiceBaseURL(import.meta.env);
  if (baseURL?.trim()) {
    return baseURL.replace(/\/$/, '');
  }
  const proxy = import.meta.env.VITE_DEV_PROXY_TARGET?.trim();
  if (proxy) {
    return proxy.replace(/\/$/, '');
  }
  return '';
}

export function fetchSysDeviceAudioList(params: { page?: number; pageSize?: number }) {
  return request<Api.Biz.DeviceAudioAdminList>(
    {
      url: '/sys/device-audio/list',
      method: 'get',
      params
    },
    { total: 0, list: [] }
  );
}

export function fetchSysDeviceAudioSpeakers() {
  return request<{ list: Api.Biz.UserSpeaker[] }>(
    { url: '/sys/device-audio/speakers', method: 'get' },
    { list: [] }
  );
}

export function fetchSysDeviceAudioBind(data: {
  deviceAudioId: string;
  speakerId?: string;
  speakerName?: string;
}) {
  return request<Api.Biz.DeviceAudioBindResult>(
    { url: '/sys/device-audio/bind', method: 'post', data },
    { deviceAudioId: '', speakerId: '', speakerName: '' }
  );
}

/** Fetch WAV blob for playback (requires auth). */
export async function fetchSysDeviceAudioSpeakerPersonal(speakerId: string) {
  const res = await requestRaw<Api.Biz.SpeakerPersonalLinesResp>({
    url: `/sys/device-audio/speakers/${encodeURIComponent(speakerId)}/personal-lines`,
    method: 'get'
  });
  if (res.error || !res.data?.speakerId) {
    throw res.error ?? new Error('load personal lines failed');
  }
  return res.data;
}

export async function fetchSysDeviceAudioSpeakerPersonalSave(speakerId: string, lines: string[]) {
  const res = await requestRaw<Api.Biz.SpeakerPersonalLinesResp>({
    url: `/sys/device-audio/speakers/${encodeURIComponent(speakerId)}/personal-lines`,
    method: 'post',
    data: { lines }
  });
  if (res.error || !res.data?.speakerId) {
    throw res.error ?? new Error('save personal lines failed');
  }
  return res.data;
}

export async function fetchSysDeviceAudioBlob(id: string): Promise<Blob> {
  const apiBase = resolveApiBase();
  const path = `/sys/device-audio/${id}/file`;
  const url = apiBase ? `${apiBase}${path}` : path;
  const res = await axios.get(url, {
    responseType: 'blob',
    headers: {
      Authorization: getAuthorization() || ''
    }
  });
  const blob = res.data as Blob;
  if (blob.type.includes('json')) {
    const text = await blob.text();
    try {
      const err = JSON.parse(text) as { message?: string };
      throw new Error(err.message || text);
    } catch {
      throw new Error(text || 'audio load failed');
    }
  }
  return blob;
}
