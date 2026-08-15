import React from 'react';
import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import { ConfigProvider, theme as antdTheme } from 'antd';
import SidebarLayout from './layouts/SidebarLayout';
import Login from './pages/Login';
import ChangePassword from './pages/ChangePassword';
import Dashboard from './pages/Dashboard';
import Config from './pages/Config';
import Logs from './pages/Logs';
import Settings from './pages/Settings';
import Core from './pages/Core';
import ListenersPage from './pages/Listeners';
import { isAuthenticated, mustChangePassword } from './api/auth';
import { ThemeProvider, useTheme } from './theme';
import './mobile.css';
import './theme.css';

const ProtectedLayout: React.FC<React.PropsWithChildren> = ({ children }) => {
  const location = useLocation();
  if (!isAuthenticated()) return <Navigate to="/login" replace state={{ from: location }} />;
  if (mustChangePassword() && location.pathname !== '/change-password') return <Navigate to="/change-password" replace />;
  return <SidebarLayout>{children}</SidebarLayout>;
};

function AppTheme({ children }: { children: React.ReactNode }) {
  const { theme } = useTheme();
  return (
    <ConfigProvider
      theme={{
        algorithm: theme === 'dark' ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
        token: {
          colorPrimary: '#1677ff',
          borderRadius: 10,
          colorBgBase: theme === 'dark' ? '#0b0f14' : '#f7f8fa',
          colorBgContainer: theme === 'dark' ? '#121820' : '#ffffff',
        },
      }}
    >
      {children}
    </ConfigProvider>
  );
}

export default function App() {
  return (
    <ThemeProvider>
      <AppTheme>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/change-password" element={<ChangePassword />} />
            <Route path="/*" element={
              <ProtectedLayout>
                <Routes>
                  <Route path="dashboard" element={<Dashboard />} />
                  <Route path="core" element={<Core />} />
                  <Route path="nodes" element={<ListenersPage />} />
                  <Route path="listeners" element={<Navigate to="/nodes" replace />} />
                  <Route path="proxies" element={<Navigate to="/nodes" replace />} />
                  <Route path="users" element={<Navigate to="/nodes" replace />} />
                  <Route path="client-access" element={<Navigate to="/nodes" replace />} />
                  <Route path="config" element={<Config />} />
                  <Route path="logs" element={<Logs />} />
                  <Route path="settings" element={<Settings />} />
                  <Route path="*" element={<Navigate to="/dashboard" replace />} />
                </Routes>
              </ProtectedLayout>
            } />
          </Routes>
        </BrowserRouter>
      </AppTheme>
    </ThemeProvider>
  );
}
