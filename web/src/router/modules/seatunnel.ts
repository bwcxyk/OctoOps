import Layout from '@/layouts/index.vue';

export default [
  {
    path: '/seatunnel',
    name: 'seatunnel',
    component: Layout,
    redirect: '/seatunnel/stream',
    meta: {
      title: { zh_CN: 'Seatunnel', en_US: 'Seatunnel' },
      icon: 'data-base',
      orderNo: 2,
    },
    children: [
      {
        path: 'stream',
        name: 'SeatunnelStreamTask',
        component: () => import('@/pages/seatunnel/stream/index.vue'),
        meta: { title: { zh_CN: '实时任务', en_US: 'Stream Task' }, permission: 'etl:stream:read' },
      },
      {
        path: 'batch',
        name: 'SeatunnelBatchTask',
        component: () => import('@/pages/seatunnel/batch/index.vue'),
        meta: { title: { zh_CN: '离线任务', en_US: 'Batch Task' }, permission: 'etl:batch:read' },
      },
    ],
  },
];
