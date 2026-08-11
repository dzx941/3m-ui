import React from 'react';
import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import SidebarLayout from './layouts/SidebarLayout';
import Login from './pages/Login';
import ChangePassword from './pages/ChangePassword';
import Dashboard from './pages/Dashboard';
import Listeners from './pages/Listeners';
import Users from './pages/Users';
import Subscriptions from './pages/Subscriptions';
import Config from './pages/Config';
import Logs from './pages/Logs';
import Settings from './pages/Settings';
import { isAuthenticated, mustChangePassword } from './api/auth';

const ProtectedLayout: React.FC<React.PropsWithChildren> = ({ children }) => {
  const location = useLocation();
  if (!isAuthenticated()) return <Navigate to="/login" replace state={{ from: location }} />;
  if (mustChangePassword() && location.pathname !== '/change-password') {
    return <Navigate to="/change-password" replace />;
  }
  return <SidebarLayout>{children}</SidebarLayout>;
};

const App: React.FC = () => {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />\n        <Route path="/change-password" element={<ChangePassword />} />
        <Route
          path="/*"
          element={
            <ProtectedLayout>
              <Routes>
                <Route path="dashboard" element={<Dashboard />} />
                <Route path="listeners" element={<Listeners />} />
                <Route path="nodes" element={<Listeners />} />
                <Route path="users" element={<Users />} />
                <Route path="subscriptions" element={<Subscriptions />} />
                <Route path="config" element={<Config />} />
                <Route path="logs" element={<Logs />} />
                <Route path="settings" element={<Settings />} />
                <Route path="*" element={<Navigate to="/dashboard" replace />} />
              </Routes>
            </ProtectedLayout>
          }
        />
      </Routes>
    </BrowserRouter>
  );
};

export default App;
