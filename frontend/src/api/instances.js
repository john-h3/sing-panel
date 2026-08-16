import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 15000
})

export const instancesApi = {
  list() {
    return api.get('/instances')
  },
  create(data) {
    return api.post('/instances', data)
  },
  update(id, data) {
    return api.put(`/instances/${id}`, data)
  },
  remove(id) {
    return api.delete(`/instances/${id}`)
  },
  checkAll() {
    return api.get('/instances/status')
  },
  checkOne(id) {
    return api.get(`/instances/${id}/status`)
  },
  sync(id, action) {
    return api.post(`/instances/${id}/sync`, { action })
  },
  syncAll() {
    return api.post('/instances/sync-all')
  },
  localInfo() {
    return api.get('/instances/local-info')
  },
  setSyncToken(token) {
    return api.put('/instances/sync-token', { token })
  },
  getConfigDiff(id) {
    return api.get(`/instances/${id}/diff`)
  }
}

export default api