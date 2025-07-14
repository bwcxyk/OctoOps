import { useUserStore } from '@/store/user'

export function hasPermission(permission) {
  const userStore = useUserStore()
  return userStore.permissions.includes(permission)
}

export function hasRole(role) {
  const userStore = useUserStore()
  return userStore.roles.includes(role)
}
