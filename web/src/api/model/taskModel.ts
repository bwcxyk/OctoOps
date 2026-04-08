export interface CustomTaskItem {
  id: number;
  name: string;
  custom_type: string;
  cron_expr: string;
  description?: string;
  status: 0 | 1;
  last_run_time?: string;
  last_result?: string;
  created_at?: string;
  updated_at?: string;
}

export interface CustomTaskListResult {
  data: CustomTaskItem[];
  total: number;
}

export interface SchedulerActiveTask {
  entry_id: number;
  task_name: string;
  task_type: 'etl' | 'custom' | string;
  next_run?: string;
}

export interface SchedulerStatusResult {
  scheduler_running: boolean;
  active_tasks_count: number;
  active_tasks: SchedulerActiveTask[];
}
