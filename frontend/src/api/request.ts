export interface ApiErrorPayload { message: string; status?: number; data?: unknown }

const TOKEN_KEY = '3m-ui.jwt';

export const getApiBaseURL = () => {
  const configured = import.meta.env.VITE_API_BASE_URL as string | undefined;
  if (configured) return configured.replace(/\/$/, '');
  if (import.meta.env.DEV) return 'http://127.0.0.1:8080/api/v1';
  return `${window.location.origin}/api/v1`;
};

export const authToken = {
  get: () => localStorage.getItem(TOKEN_KEY),
  set: (token: string) => localStorage.setItem(TOKEN_KEY, token),
  clear: () => localStorage.removeItem(TOKEN_KEY),
  key: TOKEN_KEY,
};

const normalizePath = (path: string) => path.startsWith('http') ? path : `${getApiBaseURL()}${path.startsWith('/') ? path : `/${path}`}`;

export async function apiRequest<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  if (!headers.has('Content-Type') && init.body) headers.set('Content-Type', 'application/json');
  const token = authToken.get();
  if (token && !headers.has('Authorization')) headers.set('Authorization', `Bearer ${token}`);

  let res: Response;
  try {
    res = await fetch(normalizePath(path), { ...init, headers });
  } catch {
    throw { message: 'Backend service unreachable' } satisfies ApiErrorPayload;
  }

  const contentType = res.headers.get('content-type') || '';
  const data = contentType.includes('application/json') ? await res.json().catch(() => undefined) : await res.text().catch(() => undefined);
  if (!res.ok) {
    if (res.status === 401) authToken.clear();
    const message = typeof data === 'object' && data && 'error' in data ? String((data as { error: unknown }).error) : `Request failed (${res.status})`;
    throw { message, status: res.status, data } satisfies ApiErrorPayload;
  }
  return data as T;
}

export function apiUrl(path: string) { return normalizePath(path); }
