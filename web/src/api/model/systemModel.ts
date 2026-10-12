export interface TaskSettingValue {
  enabled: boolean;
  cron: string;
}

export type SystemSettings = Record<string, TaskSettingValue | unknown>;
