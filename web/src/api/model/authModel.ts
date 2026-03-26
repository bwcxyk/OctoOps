export interface LoginRequest {
  username: string;
  password: string;
}

export interface AuthUser {
  id: number;
  username: string;
  email: string;
  nickname: string;
  avatar: string;
  status: number;
}

export interface LoginResult {
  token: string;
  user: AuthUser;
  roles: string[];
  permissions: string[];
}

export interface ProfileResult {
  user: AuthUser;
  roles: string[];
  permissions: string[];
}
