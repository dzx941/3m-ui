export interface ApiErrorPayload {
  message: string;
  status?: number;
  data?: unknown;
}

const TOKEN_KEY = '3m-ui.jwt';

export const getApiBaseURL = () => {
  const configured = import.meta.env.VITE_API_BASE_URL as string | undefined;
  if (configured) return configured.replace(/\/$/, '');
  // Production: prefer same origin. This keeps embedded SPA deployments working.
  // If the UI is opened from a different port during development, VITE_API_BASE_URL
  // can override it.
  if (typeof window !== 'undefined' && window.location?.origin) {
    return `${window.location.origin}/api/v1`;
  }
  return '/api/v1';
};

export const authToken = {
  get: () => localStorage.getItem(TOKEN_KEY),
  set: (token: string) => localStorage.setItem(TOKEN_KEY, token),
  clear: () => localStorage.removeItem(TOKEN_KEY),
  key: TOKEN_KEY,
};

const normalizePath = (path: string) => {
  if (/^https?:\/\//i.test(path)) return path;
  const base = getApiBaseURL();
  return `${base}${path.startsWith('/') ? path : `/${path}`}`;
};

async function readResponse(res: Response): Promise<unknown> {
  const contentType = res.headers.get('content-type') || '';
  if (res.status === 204) return undefined;
  if (contentType.includes('application/json')) {
    return res.json().catch(() => undefined);
  }
  return res.text().catch(() => undefined);
}

export async function apiRequest<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  if (!headers.has('Accept')) headers.set('Accept', 'application/json');
  if (!headers.has('Content-Type') && init.body && !(init.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json');
  }

  const token = authToken.get();
  if (token && !headers.has('Authorization')) {
    headers.set('Authorization', `Bearer ${token}`);
  }

  let res: Response;
  try {
    res = await fetch(normalizePath(path), {
      ...init,
      headers,
      credentials: 'same-origin',
    });
  } catch {
    throw { message: 'Backend service unreachable' } satisfies ApiErrorPayload;
  }

  const data = await readResponse(res);
  if (!res.ok) {
    if (res.status === 401) authToken.clear();

    const message =
      typeof data === 'object' && data && 'error' in data
        ? String((data as { error: unknown }).error)
        : `Request failed (${res.status})`;

    throw { message, status: res.status, data } satisfies ApiErrorPayload;
  }

  return data as T;
}

export async function downloadFile(path: string, fallbackName: string): Promise<void> {
  const headers = new Headers();
  const token = authToken.get();
  if (token) headers.set('Authorization', `Bearer ${token}`);

  let res: Response;
  try {
    res = await fetch(normalizePath(path), { headers, credentials: 'same-origin' });
  } catch {
    throw { message: 'Backend service unreachable' } satisfies ApiErrorPayload;
  }

  if (!res.ok) {
    if (res.status === 401) authToken.clear();
    throw { message: `Request failed (${res.status})`, status: res.status } satisfies ApiErrorPayload;
  }

  const blob = await res.blob();
  const disposition = res.headers.get('content-disposition') || '';
  const match = disposition.match(/filename="?([^"]+)"?/i);
  const name = match?.[1] || fallbackName;
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement('a');
  anchor.href = url;
  anchor.download = name;
  document.body.appendChild(anchor);
  anchor.click();
  anchor.remove();
  URL.revokeObjectURL(url);
}

export function apiUrl(path: string) {
  return normalizePath(path);
}
