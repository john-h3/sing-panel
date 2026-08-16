import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000
})

// Kernel API
export const kernelApi = {
  // Get current kernel status
  getStatus() {
    return api.get('/kernel/status')
  },

  // Get system info
  getSystemInfo() {
    return api.get('/kernel/system')
  },

  // Get Go runtime monitor stats
  getMonitor() {
    return api.get('/kernel/monitor')
  }
}

export default api
