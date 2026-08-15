import { apiRequest, authToken, MUST_CHANGE_KEY, USERNAME_KEY } from './request';

interface LoginResponse {
  token?: string;
  access_token?: string;
  username?: string;
  must_change_password?: boolean;
}
export interface LoginInput { username: string; password: string }

export async function login(input: LoginInput) {
  const data = await apiRequest<LoginResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify(input),
  });
  const token = data.token || data.access_token;
  if (token) authToken.set(token);
  sessionStorage.setItem(USERNAME_KEY, data.username || input.username);
  sessionStorage.setItem(MUST_CHANGE_KEY, data.must_change_password ? '1' : '0');
  return data;
}

export function logout() {
  authToken.clear();
  sessionStorage.removeItem(USERNAME_KEY);
  sessionStorage.removeItem(MUST_CHANGE_KEY);
}

export function isAuthenticated() {
  return Boolean(authToken.get());
}

export function mustChangePassword() {
  return sessionStorage.getItem(MUST_CHANGE_KEY) === '1';
}

export async function changePassword(currentPassword: string, newPassword: string) {
  const data = await apiRequest<{ status: string }>('/auth/password', {
    method: 'POST',
    body: JSON.stringify({
      current_password: currentPassword,
      new_password: newPassword,
    }),
  });
  sessionStorage.setItem(MUST_CHANGE_KEY, '0');
  return data;
}
