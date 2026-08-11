import { apiRequest, authToken } from './request';

interface LoginResponse { token?: string; access_token?: string; username?: string }
export interface LoginInput { username: string; password: string }

export async function login(input: LoginInput) {
  const data = await apiRequest<LoginResponse>('/auth/login', { method: 'POST', body: JSON.stringify(input) });
  const token = data.token || data.access_token;
  if (token) authToken.set(token);
  localStorage.setItem('3m-ui.username', data.username || input.username);
  return data;
}

export function logout() {
  authToken.clear();
  localStorage.removeItem('3m-ui.username');
}

export function isAuthenticated() { return Boolean(authToken.get()); }
