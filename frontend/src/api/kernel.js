import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 300000 // 5 minutes for downloads
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

  // Get available versions
  getVersions() {
    return api.get('/kernel/versions')
  },

  // Refresh versions cache
  refreshVersions() {
    return api.post('/kernel/versions/refresh')
  },

  // Download kernel
  download(data) {
    return api.post('/kernel/download', data)
  },

  // Stop download
  stopDownload() {
    return api.post('/kernel/stop')
  },

  // Remove kernel
  remove() {
    return api.delete('/kernel')
  },

  // Switch version
  switchVersion(version) {
    return api.post('/kernel/switch', { version })
  }
}

export default api
