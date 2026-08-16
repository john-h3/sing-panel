import axios from 'axios'

const api = axios.create({
  baseURL: '/api/process',
  timeout: 10000
})

export const processApi = {
  getStatus() {
    return api.get('/status')
  },
  getRuntimeConfig() {
    return api.get('/config')
  },
  start() {
    return api.post('/start')
  },
  stop() {
    return api.post('/stop')
  },
  restart() {
    return api.post('/restart')
  }
}

export default api
