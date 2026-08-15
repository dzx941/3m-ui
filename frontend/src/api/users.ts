import client from './client';

export interface ProxyUser {
  id: number;
  username: string;
  password?: string;
  enabled: boolean;
  created_at?: string;
}

export const fetchUsers = () => client.get<ProxyUser[]>('/users').then(r => r.data);
export const createUser = (payload: Partial<ProxyUser>) => client.post<ProxyUser>('/users', payload).then(r => r.data);
export const updateUser = (id: number, payload: Partial<ProxyUser>) => client.put<ProxyUser>(`/users/${id}`, payload).then(r => r.data);
export const deleteUser = (id: number) => client.delete(`/users/${id}`);
