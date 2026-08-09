import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import SidebarLayout from './layouts/SidebarLayout';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Listeners from './pages/Listeners';
import Users from './pages/Users';
import Subscriptions from './pages/Subscriptions';
import Config from './pages/Config';
import Logs from './pages/Logs';
import Settings from './pages/Settings';

const App: React.FC = () => {
  return (
    <BrowserRouter>
      <Routes>
        {/* Public Login Route */}
        <Route path="/login" element={<Login />} />

        {/* Dashboard and Core Management Pages Wrapped in Sidebar Layout */}
        <Route
          path="/*"
          element={
            <SidebarLayout>
              <Routes>
                <Route path="dashboard" element={<Dashboard />} />
                <Route path="listeners" element={<Listeners />} />
                <Route path="users" element={<Users />} />
                <Route path="subscriptions" element={<Subscriptions />} />
                <Route path="config" element={<Config />} />
                <Route path="logs" element={<Logs />} />
                <Route path="settings" element={<Settings />} />
                {/* Redirect any other path to dashboard */}
                <Route path="*" element={<Navigate to="/dashboard" replace />} />
              </Routes>
            </SidebarLayout>
          }
        />
      </Routes>
    </BrowserRouter>
  );
};

export default App;
