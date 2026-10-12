export interface SchedulerActiveTask {
  entry_id: number;
  task_name: string;
  task_type: 'etl' | 'system' | string;
  next_run?: string;
}

export interface SchedulerStatusResult {
  scheduler_running: boolean;
  active_tasks_count: number;
  active_tasks: SchedulerActiveTask[];
}
