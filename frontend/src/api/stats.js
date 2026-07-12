import axios from 'axios'

const api = axios.create({
  baseURL: '/api/stats',
  timeout: 10000
})

export const statsApi = {
  getServiceInfo() {
    return api.get('/service')
  }
}

export default api
