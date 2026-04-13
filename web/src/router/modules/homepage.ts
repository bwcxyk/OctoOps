import Layout from '@/layouts/index.vue';

export default [
  {
    path: '/dashboard',
    name: 'dashboard',
    component: Layout,
    redirect: '/dashboard/base',
    meta: {
      title: { zh_CN: '仪表盘', en_US: 'Dashboard' },
      icon: 'dashboard',
      orderNo: 0,
    },
    children: [
      {
        path: 'base',
        name: 'DashboardBase',
        component: () => import('@/pages/dashboard/base/index.vue'),
        meta: {
          title: { zh_CN: '基础仪表盘', en_US: 'Base Dashboard' },
          hidden: false,
        },
      },
      {
        path: 'detail',
        name: 'DashboardDetail',
        component: () => import('@/pages/dashboard/detail/index.vue'),
        meta: {
          title: { zh_CN: '详情仪表盘', en_US: 'Detail Dashboard' },
          hidden: false,
        },
      },
    ],
  },
];
