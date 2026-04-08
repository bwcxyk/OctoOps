import { request } from '@/utils/request';

import type {
  CreatePermissionPayload,
  CreateRolePayload,
  CreateUserPayload,
  PageQuery,
  PermissionListResult,
  PermissionTreeNode,
  RoleListResult,
  UpdatePermissionPayload,
  UpdateRolePayload,
  UpdateUserPayload,
  UserListResult,
} from './model/rbacModel';

interface RawResponse<T> {
  code: number;
  data: T;
  message?: string;
}

const Api = {
  Users: '/users',
  Roles: '/roles',
  Permissions: '/permissions',
  PermissionTree: '/permissions/tree',
};

async function unwrap<T>(promise: Promise<RawResponse<T>>) {
  const res = await promise;
  if (res.code !== 200) {
    throw new Error(res.message || `请求失败，错误码: ${res.code}`);
  }
  return res.data;
}

export function getUsersApi(params: PageQuery) {
  return unwrap<UserListResult>(
    request.get(
      {
        url: Api.Users,
        params,
      },
      { isTransformResponse: false },
    ),
  );
}

export function createUserApi(data: CreateUserPayload) {
  return unwrap(
    request.post(
      {
        url: Api.Users,
        data,
      },
      { isTransformResponse: false },
    ),
  );
}

export function updateUserApi(id: number, data: UpdateUserPayload) {
  return unwrap(
    request.put(
      {
        url: `${Api.Users}/${id}`,
        data,
      },
      { isTransformResponse: false },
    ),
  );
}

export function deleteUserApi(id: number) {
  return unwrap(
    request.delete(
      {
        url: `${Api.Users}/${id}`,
      },
      { isTransformResponse: false },
    ),
  );
}

export function getRolesApi(params: PageQuery) {
  return unwrap<RoleListResult>(
    request.get(
      {
        url: Api.Roles,
        params,
      },
      { isTransformResponse: false },
    ),
  );
}

export function createRoleApi(data: CreateRolePayload) {
  return unwrap(
    request.post(
      {
        url: Api.Roles,
        data,
      },
      { isTransformResponse: false },
    ),
  );
}

export function updateRoleApi(id: number, data: UpdateRolePayload) {
  return unwrap(
    request.put(
      {
        url: `${Api.Roles}/${id}`,
        data,
      },
      { isTransformResponse: false },
    ),
  );
}

export function deleteRoleApi(id: number) {
  return unwrap(
    request.delete(
      {
        url: `${Api.Roles}/${id}`,
      },
      { isTransformResponse: false },
    ),
  );
}

export function getPermissionsApi(params: PageQuery) {
  return unwrap<PermissionListResult>(
    request.get(
      {
        url: Api.Permissions,
        params,
      },
      { isTransformResponse: false },
    ),
  );
}

export function createPermissionApi(data: CreatePermissionPayload) {
  return unwrap(
    request.post(
      {
        url: Api.Permissions,
        data,
      },
      { isTransformResponse: false },
    ),
  );
}

export function updatePermissionApi(id: number, data: UpdatePermissionPayload) {
  return unwrap(
    request.put(
      {
        url: `${Api.Permissions}/${id}`,
        data,
      },
      { isTransformResponse: false },
    ),
  );
}

export function deletePermissionApi(id: number) {
  return unwrap(
    request.delete(
      {
        url: `${Api.Permissions}/${id}`,
      },
      { isTransformResponse: false },
    ),
  );
}

export function getPermissionTreeApi() {
  return unwrap<PermissionTreeNode[]>(
    request.get(
      {
        url: Api.PermissionTree,
      },
      { isTransformResponse: false },
    ),
  );
}
