import axios from 'axios'

const api = axios.create({
  baseURL: '/api/system',
  timeout: 10000
})

export const systemApi = {
  getInitSystem() {
    return api.get('/init')
  },
  restartService() {
    return api.post('/restart-service')
  },
  rebootMachine() {
    return api.post('/reboot-machine')
  }
}

export default api
