import React from 'react';
import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import SidebarLayout from './layouts/SidebarLayout';
import Login from './pages/Login';
import ChangePassword from './pages/ChangePassword';
import Dashboard from './pages/Dashboard';
import Config from './pages/Config';
import Logs from './pages/Logs';
import Settings from './pages/Settings';
import Core from './pages/Core';
import ProxyNodes from './pages/ProxyNodes';
import ListenersPage from './pages/Listeners';
import UsersPage from './pages/Users';
import ClientAccessPage from './pages/ClientAccess';
import { isAuthenticated, mustChangePassword } from './api/auth';

const ProtectedLayout: React.FC<React.PropsWithChildren> = ({ children }) => {
  const location = useLocation();
  if (!isAuthenticated()) return <Navigate to="/login" replace state={{ from: location }} />;
  if (mustChangePassword() && location.pathname !== '/change-password') {
    return <Navigate to="/change-password" replace />;
  }
  return <SidebarLayout>{children}</SidebarLayout>;
};

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/change-password" element={<ChangePassword />} />
        <Route path="/*" element = {
          <ProtectedLayout>
            <Routes>
              <Route path="dashboard" element={<Dashboard />} />
              <Route path="core" element={<Core />} />
              <Route path="proxies" element={<ProxyNodes />} />
              <Route path="listeners" element={<ListenersPage />} />
              <Route path="users" element={<UsersPage />} />
              <Route path="client-access" element={<ClientAccessPage />} />
              <Route path="config" element={<Config />} />
              <Route path="logs" element={<Logs />} />
              <Route path="settings" element={<Settings />} />
              <Route path="*" element={<Navigate to="/dashboard" replace />} />
            </Routes>
          </ProtectedLayout>
        } />
      </Routes>
    </BrowserRouter>
  );
}
