'use client';

import React from 'react';
import { usePathname, useRouter } from 'next/navigation';
import { useAuth } from './apps/web/src/context/AuthContext';
import Link from 'next/link';

const titles: Record<string, string> = {
  '/dashboard': 'Dashboard',
  '/mothers': 'Mothers',
  '/chws': 'Community Health Workers',
  '/facilities': 'Facilities',
  '/sos': 'SOS Events',
  '/visits': 'Visits',
  '/children': 'Children',
  '/notifications': 'Notifications',
  '/reports': 'Reports',
  '/settings': 'Settings',
};

export default function Header() {
  const { logout } = useAuth();
  const router = useRouter();
  const path = usePathname();
  const title = titles[path] || '';

  const handleLogout = () => {
    logout();
    router.push('/login');
  };

  return (
    <header className="flex items-center justify-between px-6 py-4 border-b border-gray-200 bg-white">
      <h1 className="text-xl font-semibold">{title}</h1>
      <div>
        <button
          onClick={handleLogout}
          className="text-sm text-red-600 hover:underline"
        >
          Logout
        </button>
      </div>
    </header>
  );
} 