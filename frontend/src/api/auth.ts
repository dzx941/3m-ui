import { apiRequest, authToken } from './request';

interface LoginResponse {
  token?: string;
  access_token?: string;
  username?: string;
  must_change_password?: boolean;
}
export interface LoginInput { username: string; password: string }

export async function login(input: LoginInput) {
  const data = await apiRequest<LoginResponse>('/auth/login', { method: 'POST', body: JSON.stringify(input) });
  const token = data.token || data.access_token;
  if (token) authToken.set(token);
  localStorage.setItem('3m-ui.username', data.username || input.username);
  localStorage.setItem('3m-ui.must_change_password', data.must_change_password ? '1' : '0');
  return data;
}

export function logout() {
  authToken.clear();
  localStorage.removeItem('3m-ui.username');
  localStorage.removeItem('3m-ui.must_change_password');
}

export function isAuthenticated() { return Boolean(authToken.get()); }


export function mustChangePassword() {
  return localStorage.getItem('3m-ui.must_change_password') === '1';
}

export async function changePassword(currentPassword: string, newPassword: string) {
  const data = await apiRequest<{ status: string }>('/auth/password', {
    method: 'POST',
    body: JSON.stringify({
      current_password: currentPassword,
      new_password: newPassword,
    }),
  });
  localStorage.setItem('3m-ui.must_change_password', '0');
  return data;
}
