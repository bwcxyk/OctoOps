import Layout from '@/layouts/index.vue';

export default [
  {
    path: '/task',
    name: 'task',
    component: Layout,
    redirect: '/task/scheduler',
    meta: {
      title: { zh_CN: '任务中心', en_US: 'Task Center' },
      icon: 'task',
      orderNo: 3,
    },
    children: [
      {
        path: 'scheduler',
        name: 'SchedulerManage',
        component: () => import('@/pages/task/scheduler/index.vue'),
        meta: { title: { zh_CN: '调度器状态', en_US: 'Scheduler Status' }, permission: 'task:scheduler' },
      },
      {
        path: 'custom',
        name: 'CustomTaskManage',
        component: () => import('@/pages/task/custom/index.vue'),
        meta: { title: { zh_CN: '自定义任务', en_US: 'Custom Task' }, permission: 'task:schedule' },
      },
      {
        path: 'log',
        name: 'TaskLogManage',
        component: () => import('@/pages/task/log/index.vue'),
        meta: { title: { zh_CN: '任务日志', en_US: 'Task Log' }, permission: 'task:log' },
      },
    ],
  },
];
