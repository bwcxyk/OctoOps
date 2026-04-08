import { request } from '@/utils/request';

import type { LoginRequest, LoginResult, ProfileResult } from './model/authModel';

interface RawResponse<T> {
  code: number;
  data: T;
  message?: string;
}

const Api = {
  Login: '/auth/login',
  Profile: '/auth/profile',
  ChangePassword: '/users/change-password',
  SendResetCode: '/users/send-reset-code',
  ForgotPassword: '/users/forgot-password',
};

async function unwrap<T>(promise: Promise<RawResponse<T>>) {
  const res = await promise;
  if (res.code !== 200) {
    throw new Error(res.message || `请求失败，错误码: ${res.code}`);
  }
  return res.data;
}

export function loginApi(data: LoginRequest) {
  return unwrap<LoginResult>(
    request.post(
      {
        url: Api.Login,
        data,
      },
      { isTransformResponse: false, withToken: false },
    ),
  );
}

export function getProfileApi() {
  return unwrap<ProfileResult>(
    request.get(
      {
        url: Api.Profile,
      },
      { isTransformResponse: false },
    ),
  );
}

export function sendResetCodeApi(email: string) {
  return unwrap(
    request.post(
      {
        url: Api.SendResetCode,
        data: { email },
      },
      { isTransformResponse: false, withToken: false },
    ),
  );
}

export function forgotPasswordApi(data: { email: string; code: string; new_password: string }) {
  return unwrap(
    request.post(
      {
        url: Api.ForgotPassword,
        data,
      },
      { isTransformResponse: false, withToken: false },
    ),
  );
}

export function changePasswordApi(data: { old_password: string; new_password: string }) {
  return unwrap(
    request.post(
      {
        url: Api.ChangePassword,
        data,
      },
      { isTransformResponse: false },
    ),
  );
}
