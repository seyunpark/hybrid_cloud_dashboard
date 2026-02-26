import { NavLink, Outlet } from 'react-router-dom';

const navItems = [
  { to: '/', label: 'Dashboard', icon: '⊞' },
  { to: '/deploy', label: 'Deployments', icon: '▶' },
  { to: '/logs', label: 'Logs', icon: '☰' },
];

export function Layout() {
  return (
    <div className="flex h-screen bg-gray-50">
      {/* Sidebar */}
      <aside className="flex w-60 flex-col border-r border-gray-200 bg-white">
        <div className="flex h-14 items-center border-b border-gray-200 px-4">
          <h1 className="text-lg font-bold text-gray-900">
            Hybrid Cloud
          </h1>
        </div>
        <nav className="flex-1 p-3">
          <ul className="space-y-1">
            {navItems.map((item) => (
              <li key={item.to}>
                <NavLink
                  to={item.to}
                  end={item.to === '/'}
                  className={({ isActive }) =>
                    `flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition-colors ${
                      isActive
                        ? 'bg-blue-50 text-blue-700'
                        : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
                    }`
                  }
                >
                  <span className="text-base">{item.icon}</span>
                  {item.label}
                </NavLink>
              </li>
            ))}
          </ul>
        </nav>
        <div className="border-t border-gray-200 p-3">
          <p className="text-xs text-gray-400">Hybrid Cloud Dashboard v0.1</p>
        </div>
      </aside>

      {/* Main Content */}
      <div className="flex flex-1 flex-col overflow-hidden">
        <header className="flex h-14 items-center justify-between border-b border-gray-200 bg-white px-6">
          <h2 className="text-sm font-medium text-gray-500">
            AI-Powered Hybrid Environment Monitoring
          </h2>
          <div className="flex items-center gap-3">
            <span className="inline-flex h-2 w-2 rounded-full bg-green-400" />
            <span className="text-xs text-gray-500">System Online</span>
          </div>
        </header>
        <main className="flex-1 overflow-auto p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
