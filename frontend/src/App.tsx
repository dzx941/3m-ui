import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, App as AntApp } from 'antd';
import AppLayout from './components/AppLayout';
import ProtectedRoute from './components/ProtectedRoute';
import Login from './pages/Login';
import ChangePassword from './pages/ChangePassword';
import Dashboard from './pages/Dashboard';
import Listeners from './pages/Listeners';
import Users from './pages/Users';
import Core from './pages/Core';
import Logs from './pages/Logs';
import ConfigPage from './pages/Config';
import Settings from './pages/Settings';
import InboundTemplates from './pages/InboundTemplates';

const App: React.FC = () => {
  return (
    <ConfigProvider theme={{ token: { colorPrimary: '#1677ff' } }}>
      <AntApp>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/change-password" element={<ProtectedRoute allowUnchanged><ChangePassword /></ProtectedRoute>} />
            <Route path="/*" element={<ProtectedRoute><AppLayout><Routes>
              <Route path="/" element={<Dashboard />} />
              <Route path="/listeners" element={<Listeners />} />
              <Route path="/inbound-templates" element={<InboundTemplates />} />
              <Route path="/users" element={<Users />} />
              <Route path="/core" element={<Core />} />
              <Route path="/logs" element={<Logs />} />
              <Route path="/config" element={<ConfigPage />} />
              <Route path="/settings" element={<Settings />} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes></AppLayout></ProtectedRoute>} />
          </Routes>
        </BrowserRouter>
      </AntApp>
    </ConfigProvider>
  );
};

export default App;
