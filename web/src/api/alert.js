import axios from 'axios'

export function getAlerts() {
  return axios.get('/api/channels')
}

export function createAlert(data) {
  return axios.post('/api/channels', data)
}

export function updateAlert(id, data) {
  return axios.put(`/api/channels/${id}`, data)
}

export function deleteAlert(id) {
  return axios.delete(`/api/channels/${id}`)
} 