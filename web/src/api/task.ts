import { request } from '@/utils/request';

import type { CustomTaskListResult, SchedulerStatusResult } from './model/taskModel';

const Api = {
  CustomTasks: '/custom-tasks',
  SchedulerStatus: '/scheduler/status',
  SchedulerReload: '/scheduler/reload',
  SchedulerStart: '/scheduler/start',
  SchedulerStop: '/scheduler/stop',
  TaskLogs: '/task-logs',
};

export function getCustomTasksApi(params: { page: number; page_size: number }) {
  return request.get<CustomTaskListResult>(
    { url: Api.CustomTasks, params },
    { isTransformResponse: false },
  );
}

export function updateCustomTaskApi(id: number, data: Record<string, unknown>) {
  return request.put(
    { url: `${Api.CustomTasks}/${id}`, data },
    { isTransformResponse: false },
  );
}

export function getSchedulerStatusApi() {
  return request.get<SchedulerStatusResult>(
    { url: Api.SchedulerStatus },
    { isTransformResponse: false },
  );
}

export function reloadSchedulerApi() {
  return request.post(
    { url: Api.SchedulerReload, data: {} },
    { isTransformResponse: false },
  );
}

export function startSchedulerApi() {
  return request.post(
    { url: Api.SchedulerStart, data: {} },
    { isTransformResponse: false },
  );
}

export function stopSchedulerApi() {
  return request.post(
    { url: Api.SchedulerStop, data: {} },
    { isTransformResponse: false },
  );
}
