import Layout from '@/layouts/index.vue';

export default [
  {
    path: '/alert',
    name: 'alert',
    component: Layout,
    redirect: '/alert/group',
    meta: {
      title: { zh_CN: '告警管理', en_US: 'Alert' },
      icon: 'notification',
      orderNo: 4,
    },
    children: [
      {
        path: 'group',
        name: 'AlertGroup',
        component: () => import('@/pages/alert/group/index.vue'),
        meta: { title: { zh_CN: '告警组', en_US: 'Alert Group' }, permission: 'notify:group' },
      },
      {
        path: 'template',
        name: 'AlertTemplate',
        component: () => import('@/pages/alert/template/index.vue'),
        meta: { title: { zh_CN: '告警模板', en_US: 'Alert Template' }, permission: 'notify:template' },
      },
      {
        path: 'channel',
        name: 'AlertChannel',
        component: () => import('@/pages/alert/channel/index.vue'),
        meta: { title: { zh_CN: '告警渠道', en_US: 'Alert Channel' }, permission: 'notify:channel' },
      },
    ],
  },
];
