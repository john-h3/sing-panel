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
  }
}

export default api
