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
        meta: { title: { zh_CN: '实时任务', en_US: 'Stream Task' }, permission: 'etl:stream' },
      },
      {
        path: 'stream/create',
        name: 'SeatunnelStreamTaskCreate',
        component: () => import('@/pages/seatunnel/stream/edit.vue'),
        meta: {
          title: { zh_CN: '新增实时任务', en_US: 'Create Stream Task' },
          permission: 'etl:stream:create',
          hidden: true,
          keepAlive: false,
          activeMenu: '/seatunnel/stream',
        },
      },
      {
        path: 'stream/:id/edit',
        name: 'SeatunnelStreamTaskEdit',
        component: () => import('@/pages/seatunnel/stream/edit.vue'),
        meta: {
          title: { zh_CN: '编辑实时任务', en_US: 'Edit Stream Task' },
          permission: 'etl:stream:update',
          hidden: true,
          keepAlive: false,
          activeMenu: '/seatunnel/stream',
        },
      },
      {
        path: 'batch',
        name: 'SeatunnelBatchTask',
        component: () => import('@/pages/seatunnel/batch/index.vue'),
        meta: { title: { zh_CN: '离线任务', en_US: 'Batch Task' }, permission: 'etl:batch' },
      },
      {
        path: 'batch/create',
        name: 'SeatunnelBatchTaskCreate',
        component: () => import('@/pages/seatunnel/batch/edit.vue'),
        meta: {
          title: { zh_CN: '新增离线任务', en_US: 'Create Batch Task' },
          permission: 'etl:batch:create',
          hidden: true,
          keepAlive: false,
          activeMenu: '/seatunnel/batch',
        },
      },
      {
        path: 'batch/:id/edit',
        name: 'SeatunnelBatchTaskEdit',
        component: () => import('@/pages/seatunnel/batch/edit.vue'),
        meta: {
          title: { zh_CN: '编辑离线任务', en_US: 'Edit Batch Task' },
          permission: 'etl:batch:update',
          hidden: true,
          keepAlive: false,
          activeMenu: '/seatunnel/batch',
        },
      },
    ],
  },
];
