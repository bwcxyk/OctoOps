import axios from 'axios'

// 获取所有安全组配置
export function getAliyunSGConfigs() {
  return axios.get('/api/aliyun-sg-configs')
}

// 新增安全组配置
export function createAliyunSGConfig(data) {
  return axios.post('/api/aliyun-sg-configs', data)
}

// 更新安全组配置
export function updateAliyunSGConfig(id, data) {
  return axios.put(`/api/aliyun-sg-configs/${id}`, data)
}

// 删除安全组配置
export function deleteAliyunSGConfig(id) {
  return axios.delete(`/api/aliyun-sg-configs/${id}`)
}

// 获取IP历史
export function getAliyunIPHistory() {
  return axios.get('/api/aliyun-ip-history')
}

// 同步安全组配置
export function syncAliyunSGConfigs() {
  return axios.post('/api/aliyun-sg-configs/sync')
}

// 单条同步安全组配置
export function syncAliyunSGConfig(id) {
  return axios.post(`/api/aliyun-sg-configs/${id}/sync`)
} 