import { request } from '@/utils/request';

import type {
  SeatunnelTask,
  TaskListQuery,
  TaskListResult,
  TaskLogListQuery,
  TaskLogListResult,
} from './model/seatunnelModel';

const Api = {
  StreamTasks: '/seatunnel/stream',
  BatchTasks: '/seatunnel/batch',
  StartTask: '/seatunnel/tasks',
  StopTask: '/seatunnel/tasks',
  SyncJobStatus: '/seatunnel/tasks/sync-status',
  TaskLogs: '/task/log',
};

function resolveTasksApiByType(taskType?: string) {
  if (taskType === 'stream') return Api.StreamTasks;
  if (taskType === 'batch') return Api.BatchTasks;
  return Api.StreamTasks;
}

export function getTasksApi(params: TaskListQuery) {
  return request.get<TaskListResult>(
    {
      url: resolveTasksApiByType(params.task_type),
      params,
    },
    { isTransformResponse: false },
  );
}

export function createTaskApi(data: Partial<SeatunnelTask>) {
  return request.post<SeatunnelTask>(
    {
      url: resolveTasksApiByType(data.task_type),
      data,
    },
    { isTransformResponse: false },
  );
}

export function updateTaskApi(id: number, data: Partial<SeatunnelTask>, taskType?: 'stream' | 'batch') {
  return request.put<SeatunnelTask>(
    {
      url: `${resolveTasksApiByType(taskType)}/${id}`,
      data,
    },
    { isTransformResponse: false },
  );
}

export function deleteTaskApi(id: number, taskType?: 'stream' | 'batch') {
  return request.delete<{ message: string }>(
    {
      url: `${resolveTasksApiByType(taskType)}/${id}`,
    },
    { isTransformResponse: false },
  );
}

export function submitJobApi(params: { id: number; isStartWithSavePoint?: boolean }) {
  const query = new URLSearchParams();
  if (typeof params.isStartWithSavePoint === 'boolean') {
    query.set('isStartWithSavePoint', String(params.isStartWithSavePoint));
  }
  const queryString = query.toString();
  return request.post<{ message?: string; error?: string }>(
    {
      url: `${Api.StartTask}/${params.id}/start${queryString ? `?${queryString}` : ''}`,
      data: {},
    },
    { isTransformResponse: false },
  );
}

export function stopJobApi(params: { id: number; isStopWithSavePoint?: boolean }) {
  const query = new URLSearchParams();
  if (typeof params.isStopWithSavePoint === 'boolean') {
    query.set('isStopWithSavePoint', String(params.isStopWithSavePoint));
  }
  const queryString = query.toString();
  return request.post<{ message?: string; error?: string }>(
    {
      url: `${Api.StopTask}/${params.id}/stop${queryString ? `?${queryString}` : ''}`,
      data: {},
    },
    { isTransformResponse: false },
  );
}

export function syncJobStatusApi() {
  return request.post<{ message: string }>(
    {
      url: Api.SyncJobStatus,
      data: {},
    },
    { isTransformResponse: false },
  );
}

export function getTaskLogsApi(params: TaskLogListQuery) {
  return request.get<TaskLogListResult>(
    {
      url: Api.TaskLogs,
      params,
    },
    { isTransformResponse: false },
  );
}
