export type TaskType = 'stream' | 'batch' | string;
export type StatusType = 0 | 1 | number;

export interface SeatunnelTask {
  id: number;
  name: string;
  description?: string;
  task_type: TaskType;
  cron_expr?: string;
  config: string;
  config_format?: 'json' | 'hocon' | string;
  job_id?: string | null;
  job_status?: string;
  alert_group?: string;
  status?: StatusType;
  last_run_time?: string;
  finish_time?: string;
  next_run_time?: string;
  created_at?: string;
  updated_at?: string;
}

export interface TaskListQuery {
  task_type: TaskType;
  page: number;
  page_size: number;
  name?: string;
  status?: string | number;
  job_id?: string;
  job_status?: string;
}

export interface TaskListResult {
  data: SeatunnelTask[];
  total: number;
}

export interface TaskLogItem {
  id: number;
  task_name: string;
  status: string;
  message?: string;
  created_at?: string;
}

export interface TaskLogListQuery {
  page: number;
  page_size: number;
  task_name?: string;
  status?: string;
  start_time?: string;
  end_time?: string;
}

export interface TaskLogListResult {
  data: TaskLogItem[];
  total: number;
}
