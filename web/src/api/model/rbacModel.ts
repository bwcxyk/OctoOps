export type StatusType = 0 | 1;

export interface RoleItem {
  id: number;
  name: string;
  description: string;
  status: StatusType;
  created_at?: string;
  updated_at?: string;
  permissions?: PermissionItem[];
}

export interface PermissionItem {
  id: number;
  name: string;
  code: string;
  description: string;
  type: 'menu' | 'api' | string;
  path: string;
  method: string;
  status: StatusType;
  parent_id?: number;
  created_at?: string;
  updated_at?: string;
}

export interface PermissionTreeNode extends PermissionItem {
  children: PermissionTreeNode[];
}

export interface UserItem {
  id: number;
  username: string;
  email: string;
  nickname: string;
  avatar?: string;
  status: StatusType;
  created_at?: string;
  updated_at?: string;
  roles?: RoleItem[];
}

export interface PageQuery {
  page: number;
  page_size: number;
}

export interface UserListResult {
  users: UserItem[];
  total: number;
  page: number;
  page_size: number;
  total_page: number;
}

export interface RoleListResult {
  roles: RoleItem[];
  total: number;
  page: number;
  page_size: number;
  total_page: number;
}

export interface PermissionListResult {
  permissions: PermissionItem[];
  total: number;
  page: number;
  page_size: number;
  total_page: number;
}

export interface CreateUserPayload {
  username: string;
  password: string;
  email: string;
  nickname: string;
  role_ids: number[];
}

export interface UpdateUserPayload {
  email?: string;
  nickname?: string;
  avatar?: string;
  status?: StatusType;
  role_ids?: number[];
}

export interface CreateRolePayload {
  name: string;
  description: string;
  permission_ids: number[];
}

export interface UpdateRolePayload {
  name?: string;
  description?: string;
  status?: StatusType;
  permission_ids?: number[];
}

export interface CreatePermissionPayload {
  name: string;
  code: string;
  description: string;
  type: string;
  path: string;
  method: string;
}

export interface UpdatePermissionPayload {
  name?: string;
  code?: string;
  description?: string;
  type?: string;
  path?: string;
  method?: string;
  status?: StatusType;
}
