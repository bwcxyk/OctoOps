import { request } from '@/utils/request';

import type { SystemSettings } from './model/systemModel';

const Api = {
  SystemSettings: '/system/settings',
};

export function getSystemSettingsApi() {
  return request.get<{ data: SystemSettings }>({ url: Api.SystemSettings }, { isTransformResponse: false });
}

export function updateSystemSettingsApi(data: Record<string, unknown>) {
  return request.put<{ data: SystemSettings }>({ url: Api.SystemSettings, data }, { isTransformResponse: false });
}
