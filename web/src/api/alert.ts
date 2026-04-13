import { request } from '@/utils/request';

import type { AlertChannel, AlertGroup, AlertGroupMember, AlertTemplate } from './model/alertModel';

const Api = {
  Channels: '/channels',
  AlertGroups: '/alert-groups',
  AlertTemplates: '/alert-templates',
};

export function getAlertChannelsApi() {
  return request.get<AlertChannel[]>(
    { url: Api.Channels },
    { isTransformResponse: false },
  );
}

export function createAlertChannelApi(data: Partial<AlertChannel>) {
  return request.post<AlertChannel>(
    { url: Api.Channels, data },
    { isTransformResponse: false },
  );
}

export function updateAlertChannelApi(id: number, data: Partial<AlertChannel>) {
  return request.put<AlertChannel>(
    { url: `${Api.Channels}/${id}`, data },
    { isTransformResponse: false },
  );
}

export function deleteAlertChannelApi(id: number) {
  return request.delete<{ message: string }>(
    { url: `${Api.Channels}/${id}` },
    { isTransformResponse: false },
  );
}

export function testAlertChannelApi(id: number, template_content?: string) {
  return request.post<{ message: string; error?: string }>(
    {
      url: `${Api.Channels}/${id}/test`,
      data: { template_content: template_content || '' },
    },
    { isTransformResponse: false },
  );
}

export function getAlertGroupsApi() {
  return request.get<AlertGroup[]>(
    { url: Api.AlertGroups },
    { isTransformResponse: false },
  );
}

export function createAlertGroupApi(data: Partial<AlertGroup>) {
  return request.post<AlertGroup>(
    { url: Api.AlertGroups, data },
    { isTransformResponse: false },
  );
}

export function updateAlertGroupApi(id: number, data: Partial<AlertGroup>) {
  return request.put<AlertGroup>(
    { url: `${Api.AlertGroups}/${id}`, data },
    { isTransformResponse: false },
  );
}

export function deleteAlertGroupApi(id: number) {
  return request.delete<{ message: string }>(
    { url: `${Api.AlertGroups}/${id}` },
    { isTransformResponse: false },
  );
}

export function getAlertGroupMembersApi(groupId: number) {
  return request.get<AlertGroupMember[]>(
    { url: `${Api.AlertGroups}/${groupId}/members` },
    { isTransformResponse: false },
  );
}

export function addAlertGroupMemberApi(groupId: number, data: Pick<AlertGroupMember, 'channel_type' | 'channel_id'>) {
  return request.post<AlertGroupMember>(
    { url: `${Api.AlertGroups}/${groupId}/members`, data },
    { isTransformResponse: false },
  );
}

export function deleteAlertGroupMemberApi(groupId: number, memberId: number) {
  return request.delete<{ message: string }>(
    { url: `${Api.AlertGroups}/${groupId}/members/${memberId}` },
    { isTransformResponse: false },
  );
}

export function getAlertTemplatesApi() {
  return request.get<AlertTemplate[]>(
    { url: Api.AlertTemplates },
    { isTransformResponse: false },
  );
}

export function createAlertTemplateApi(data: Partial<AlertTemplate>) {
  return request.post<AlertTemplate>(
    { url: Api.AlertTemplates, data },
    { isTransformResponse: false },
  );
}

export function updateAlertTemplateApi(id: number, data: Partial<AlertTemplate>) {
  return request.put<AlertTemplate>(
    { url: `${Api.AlertTemplates}/${id}`, data },
    { isTransformResponse: false },
  );
}

export function deleteAlertTemplateApi(id: number) {
  return request.delete<{ message: string }>(
    { url: `${Api.AlertTemplates}/${id}` },
    { isTransformResponse: false },
  );
}
