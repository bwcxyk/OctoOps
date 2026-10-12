import Layout from '@/layouts/index.vue';

export default [
  {
    path: '/system',
    name: 'system',
    component: Layout,
    redirect: '/system/setting',
    meta: {
      title: { zh_CN: '系统设置', en_US: 'System Setting' },
      icon: 'setting',
      orderNo: 6,
    },
    children: [
      {
        path: 'setting',
        name: 'SystemSetting',
        component: () => import('@/pages/system/setting/index.vue'),
        meta: { title: { zh_CN: '基本设置', en_US: 'Basic Setting' }, permission: 'system:setting' },
      },
    ],
  },
];
