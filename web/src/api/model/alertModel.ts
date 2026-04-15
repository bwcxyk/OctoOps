export interface AlertChannel {
  id: number;
  name: string;
  type: 'email' | 'dingtalk' | 'wechat' | 'feishu' | string;
  target: string;
  status: 0 | 1;
  dingtalk_secret?: string;
  template_id?: number | null;
  created_at?: string;
  updated_at?: string;
}

export interface AlertGroup {
  id: number;
  name: string;
  description?: string;
  status: 0 | 1;
  created_at?: string;
  updated_at?: string;
}

export interface AlertTemplate {
  id: number;
  name: string;
  content: string;
  created_at?: string;
  updated_at?: string;
}

export interface AlertGroupMember {
  id: number;
  group_id: number;
  channel_type: 'email' | 'dingtalk' | 'wechat' | 'feishu' | string;
  channel_id: number;
}
