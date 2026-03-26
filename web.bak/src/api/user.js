import axios from 'axios'

export function loginApi(data) {
  return axios.post('/api/auth/login', data).then(res => res.data)
}

export function getProfileApi() {
  return axios.get('/api/auth/profile').then(res => res.data)
}

// 用户管理
export function getUsersApi(params) {
  return axios.get('/api/users', { params }).then(res => res.data)
}
export function createUserApi(data) {
  return axios.post('/api/users', data).then(res => res.data)
}
export function updateUserApi(id, data) {
  return axios.put(`/api/users/${id}`, data).then(res => res.data)
}
export function deleteUserApi(id) {
  return axios.delete(`/api/users/${id}`).then(res => res.data)
}

// 角色管理
export function getRolesApi(params) {
  return axios.get('/api/roles', { params }).then(res => res.data)
}
export function createRoleApi(data) {
  return axios.post('/api/roles', data).then(res => res.data)
}
export function updateRoleApi(id, data) {
  return axios.put(`/api/roles/${id}`, data).then(res => res.data)
}
export function deleteRoleApi(id) {
  return axios.delete(`/api/roles/${id}`).then(res => res.data)
}

// 权限管理
export function getPermissionsApi(params) {
  return axios.get('/api/permissions', { params }).then(res => res.data)
}
export function createPermissionApi(data) {
  return axios.post('/api/permissions', data).then(res => res.data)
}
export function updatePermissionApi(id, data) {
  return axios.put(`/api/permissions/${id}`, data).then(res => res.data)
}
export function deletePermissionApi(id) {
  return axios.delete(`/api/permissions/${id}`).then(res => res.data)
}

export function getPermissionTreeApi() {
  return axios.get('/api/permissions/tree').then(res => res.data)
}

export function getMenusApi() {
  return axios.get('/api/menus')
}

// 修改密码
export function changePassword(data) {
  return axios.post('/api/users/change-password', data)
}

// 忘记密码
export function forgotPassword(data) {
  return axios.post('/api/users/forgot-password', {
    email: data.email,
    code: data.code,
    new_password: data.new_password
  })
}

// 发送邮箱验证码
export function sendResetCode(data) {
  return axios.post('/api/users/send-reset-code', { email: data.email })
}
