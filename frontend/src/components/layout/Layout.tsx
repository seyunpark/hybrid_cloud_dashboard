import { useEffect, useState } from 'react';
import { Link, NavLink, Outlet } from 'react-router-dom';

const CURRENT_VERSION = '1.0.0';
const GITHUB_REPO = 'seyunpark/hybrid_cloud_dashboard';

function useVersionCheck() {
  const [latestVersion, setLatestVersion] = useState<string | null>(null);
  const [releaseUrl, setReleaseUrl] = useState<string | null>(null);

  useEffect(() => {
    fetch(`https://api.github.com/repos/${GITHUB_REPO}/releases/latest`)
      .then((res) => {
        if (!res.ok) return null;
        return res.json();
      })
      .then((data) => {
        if (!data?.tag_name) return;
        const version = data.tag_name.replace(/^v/, '');
        if (version !== CURRENT_VERSION) {
          setLatestVersion(version);
          setReleaseUrl(data.html_url);
        }
      })
      .catch(() => {});
  }, []);

  return { latestVersion, releaseUrl };
}

const navItems = [
  { to: '/', label: 'Dashboard', icon: '⊞' },
  { to: '/deploy', label: 'Deployments', icon: '▶' },
  { to: '/history', label: 'History', icon: '☰' },
  { to: '/settings', label: 'Settings', icon: '⚙' },
];

export function Layout() {
  const { latestVersion, releaseUrl } = useVersionCheck();

  return (
    <div className="flex h-screen bg-gray-50">
      {/* Sidebar */}
      <aside className="flex w-60 flex-col border-r border-gray-200 bg-white">
        <div className="flex h-14 items-center border-b border-gray-200 px-4">
          <Link to="/" className="text-lg font-bold text-gray-900 hover:text-blue-700 transition-colors">
            Vineyard
          </Link>
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
          <p className="text-xs text-gray-400">Vineyard v{CURRENT_VERSION}</p>
          {latestVersion && releaseUrl && (
            <a
              href={releaseUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="mt-1 block rounded bg-blue-50 px-2 py-1 text-xs text-blue-600 hover:bg-blue-100 transition-colors"
            >
              v{latestVersion} 업데이트 available
            </a>
          )}
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
