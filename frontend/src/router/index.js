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
      }
    ]
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('../views/Settings.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
