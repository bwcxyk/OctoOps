import Layout from '@/layouts/index.vue';

export default [
  {
    path: '/aliyun',
    name: 'aliyun',
    component: Layout,
    redirect: '/aliyun/ecs-security-group',
    meta: {
      title: { zh_CN: '阿里云', en_US: 'Aliyun' },
      icon: 'cloud',
      orderNo: 1,
    },
    children: [
      {
        path: 'ecs-security-group',
        name: 'AliyunEcsSecurityGroup',
        component: () => import('@/pages/aliyun/ecs-security-group/index.vue'),
        meta: { title: { zh_CN: '安全组同步', en_US: 'ECS Security Group' }, permission: 'aliyun:ecs_sg:read' },
      },
    ],
  },
];
