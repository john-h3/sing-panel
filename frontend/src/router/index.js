import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'Dashboard',
    component: () => import('../views/Dashboard.vue')
  },
  {
    path: '/kernel',
    name: 'Kernel',
    component: () => import('../views/System.vue')
  },
  {
    path: '/singbox',
    name: 'SingBox',
    component: () => import('../views/singbox/Config.vue'),
    redirect: '/singbox/inbounds',
    children: [
      {
        path: 'inbounds',
        name: 'Inbounds',
        component: () => import('../views/singbox/Inbounds.vue')
      },
      {
        path: 'outbounds',
        name: 'Outbounds',
        component: () => import('../views/singbox/Outbounds.vue')
      },
      {
        path: 'rulesets',
        name: 'Rulesets',
        component: () => import('../views/singbox/Rulesets.vue')
      },
      {
        path: 'route-rules',
        name: 'RouteRules',
        component: () => import('../views/singbox/RouteRules.vue')
      },
      {
        path: 'dns',
        name: 'DNS',
        component: () => import('../views/singbox/DNS.vue')
      },
      {
        path: 'services',
        name: 'Services',
        component: () => import('../views/singbox/Services.vue')
      },
      {
        path: 'http-clients',
        name: 'HTTPClients',
        component: () => import('../views/singbox/HTTPClients.vue')
      },
      {
        path: 'experimental',
        name: 'Experimental',
        component: () => import('../views/singbox/Experimental.vue')
      }
    ]
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('../views/Settings.vue'),
    redirect: '/settings/accelerate',
    children: [
      {
        path: 'accelerate',
        name: 'SettingsAccelerate',
        component: () => import('../views/settings/Accelerate.vue')
      },
      {
        path: 'dashboard',
        name: 'SettingsDashboard',
        component: () => import('../views/settings/Dashboard.vue')
      },
      {
        path: 'data',
        name: 'SettingsData',
        component: () => import('../views/settings/Data.vue')
      },
      {
        path: 'instances',
        name: 'SettingsInstances',
        component: () => import('../views/settings/Instances.vue')
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
