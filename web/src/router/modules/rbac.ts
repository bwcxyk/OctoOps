import Layout from '@/layouts/index.vue';

export default [
  {
    path: '/rbac',
    name: 'rbac',
    component: Layout,
    redirect: '/rbac/user',
    meta: {
      title: { zh_CN: '系统管理', en_US: 'RBAC' },
      icon: 'secured',
      orderNo: 5,
    },
    children: [
      {
        path: 'permission',
        name: 'RbacPermissionManage',
        component: () => import('@/pages/rbac/permission/index.vue'),
        meta: { title: { zh_CN: '权限管理', en_US: 'Permission Manage' }, permission: 'rbac:permission' },
      },
      {
        path: 'role',
        name: 'RbacRoleManage',
        component: () => import('@/pages/rbac/role/index.vue'),
        meta: { title: { zh_CN: '角色管理', en_US: 'Role Manage' }, permission: 'rbac:role' },
      },
      {
        path: 'user',
        name: 'RbacUserManage',
        component: () => import('@/pages/rbac/user/index.vue'),
        meta: { title: { zh_CN: '用户管理', en_US: 'User Manage' }, permission: 'rbac:user' },
      },
    ],
  },
];
