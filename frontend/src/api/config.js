import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000
})

export const configApi = {
  get() {
    return api.get('/config')
  },
  update(data) {
    return api.put('/config', data)
  }
}

export default api
