import client from './client';

export interface SystemStatus {
  cpu: { percent: number };
  memory: { used: number; total: number; percent: number };
  disk: { used: number; total: number; percent: number };
  network: { upload: number; download: number };
}

export interface MihomoStatus {
  running: boolean;
  version: string;
  pid: number;
  uptime: string;
}

export const fetchDashboard = () => client.get<SystemStatus & { mihomo: MihomoStatus }>('/dashboard').then(r => r.data);
export const startMihomo = () => client.post('/mihomo/start');
export const stopMihomo = () => client.post('/mihomo/stop');
export const restartMihomo = () => client.post('/mihomo/restart');
export const fetchLogs = () => client.get<{ logs: string[] }>('/mihomo/logs').then(r => r.data);
