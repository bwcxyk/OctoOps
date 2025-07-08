import axios from 'axios';

const api = axios.create({
  baseURL: '/api'
});

export function getTasks(taskType, params = {}) {
  return axios.get('/api/tasks', { params: { task_type: taskType, ...params } })
}

export function createTask(data) {
  return api.post('/tasks', data);
}

export function updateTask(id, data) {
  return api.put(`/tasks/${id}`, data);
}

export function deleteTask(id) {
  return api.delete(`/tasks/${id}`);
}

export function submitJob(data, config = {}) {
  return api.post('/submit-job', data, config);
}

export function stopJob(data, config = {}) {
  return api.post('/stop-job', data, config);
} 