export interface AliyunSgConfig {
  id: number;
  name: string;
  access_key: string;
  access_secret?: string;
  region_id: string;
  security_group_id: string;
  port_list: string;
  status: 0 | 1;
  last_ip?: string;
  last_ip_updated_at?: string;
  created_at?: string;
  updated_at?: string;
}
