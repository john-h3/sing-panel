import axios from 'axios'

const api = axios.create({
  baseURL: '/clash_api',
  timeout: 10000
})

export const clashApi = {
  // Traffic (streaming - use fetch for streaming support)
  getTraffic() {
    return fetch('/clash_api/traffic')
  },

  // Connections
  getConnections() {
    return api.get('/connections')
  },

  // Proxies
  getProxies() {
    return api.get('/proxies')
  },

  // Rules
  getRules() {
    return api.get('/rules')
  },

  // Providers
  getProviders() {
    return api.get('/providers')
  },

  // Config
  getConfig() {
    return api.get('/configs')
  },

  // WebSocket for real-time traffic
  connectTraffic() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    return new WebSocket(`${protocol}//${window.location.host}/clash_api/traffic`)
  }
}

export default api
