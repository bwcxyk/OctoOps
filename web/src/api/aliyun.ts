import { request } from '@/utils/request';

import type { AliyunSgConfig } from './model/aliyunModel';

const Api = {
  AliyunSgConfigs: '/aliyun-sg-configs',
};

export function getAliyunSgConfigsApi(params?: { name?: string; access_key?: string; status?: number }) {
  return request.get<AliyunSgConfig[]>(
    { url: Api.AliyunSgConfigs, params },
    { isTransformResponse: false },
  );
}

export function createAliyunSgConfigApi(data: Partial<AliyunSgConfig>) {
  return request.post<AliyunSgConfig>(
    { url: Api.AliyunSgConfigs, data },
    { isTransformResponse: false },
  );
}

export function updateAliyunSgConfigApi(id: number, data: Partial<AliyunSgConfig>) {
  return request.put<AliyunSgConfig>(
    { url: `${Api.AliyunSgConfigs}/${id}`, data },
    { isTransformResponse: false },
  );
}

export function deleteAliyunSgConfigApi(id: number) {
  return request.delete<{ message: string }>(
    { url: `${Api.AliyunSgConfigs}/${id}` },
    { isTransformResponse: false },
  );
}

export function syncAliyunSgConfigApi(id: number) {
  return request.post<{ message: string }>(
    { url: `${Api.AliyunSgConfigs}/${id}/sync`, data: {} },
    { isTransformResponse: false },
  );
}
