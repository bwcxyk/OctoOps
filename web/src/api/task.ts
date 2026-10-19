import { request } from '@/utils/request';

import type { SchedulerStatusResult } from './model/taskModel';

const Api = {
  SchedulerStatus: '/task/scheduler/status',
  SchedulerReload: '/task/scheduler/reload',
  SchedulerStart: '/task/scheduler/start',
  SchedulerStop: '/task/scheduler/stop',
  TaskLogs: '/task/log',
};

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
