import axios from 'axios'

export function getAlerts() {
  return axios.get('/api/alerts')
}

export function createAlert(data) {
  return axios.post('/api/alerts', data)
}

export function updateAlert(id, data) {
  return axios.put(`/api/alerts/${id}`, data)
}

export function deleteAlert(id) {
  return axios.delete(`/api/alerts/${id}`)
} 