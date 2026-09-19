import "./app.css";
import {
  isRouteErrorResponse,
  Link,
  Links,
  Meta,
  Outlet,
  useLocation,
  useNavigate,
} from "react-router";
import { AuthProvider, useAuth } from "./providers/auth";
import { VaultProvider, useVault } from "./providers/vault";
import NavLink from "./components/NavLink";
import type { Route } from "./+types/root";
import { ToastProvider } from "./providers/toast";
import ToastContainer from "./components/ToastContainer";

export const links: Route.LinksFunction = () => [
  { rel: "preconnect", href: "https://fonts.googleapis.com" },
  {
    rel: "preconnect",
    href: "https://fonts.gstatic.com",
    crossOrigin: "anonymous",
  },
  {
    rel: "stylesheet",
    href: "https://fonts.googleapis.com/css2?family=Inter:ital,opsz,wght@0,14..32,100..900;1,14..32,100..900&display=swap",
  },
];

export default function Root() {
  return (
    <AuthProvider>
      <VaultProvider>
        <ToastProvider>
          <Layout />
        </ToastProvider>
      </VaultProvider>
    </AuthProvider>
  );
}

function Layout() {
  const { user, isAuthenticated, logout } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate("/");
  };

  const isAuthPage =
    location.pathname === "/login" || location.pathname === "/signup";

  return (
    <html lang="en">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <Meta />
        <Links />
      </head>
      <body>
        <div className="min-h-screen bg-slate-950 text-slate-200 font-sans antialiased">
          <header className="sticky top-0 z-50 border-b border-slate-800 bg-slate-950">
            <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
              <Logo />
              {isAuthenticated && <Nav />}
              <Actions
                isAuthenticated={isAuthenticated}
                userEmail={user?.email}
                isAuthPage={isAuthPage}
                onLogout={handleLogout}
              />
            </div>
          </header>

          <main className="mx-auto max-w-6xl px-6 py-8">
            <Outlet />
          </main>
          <ToastContainer />
        </div>
      </body>
    </html>
  );
}

function Logo() {
  return (
    <Link to="/" className="flex items-center gap-2.5 text-sky-400">
      <svg
        width="24"
        height="24"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
        <path d="M7 11V7a5 5 0 0 1 10 0v4" />
      </svg>
      <span className="text-lg font-bold tracking-tight">LockKeep</span>
    </Link>
  );
}

function Nav() {
  const { user } = useAuth();
  
  return (
    <nav className="flex gap-1">
      <NavLink to="/dashboard">Dashboard</NavLink>
      <NavLink to="/settings">Settings</NavLink>
      {user?.userType === "admin" && (
        <NavLink to="/admin/policy">Policy</NavLink>
      )}
    </nav>
  );
}

function Actions({
  isAuthenticated,
  userEmail,
  isAuthPage,
  onLogout,
}: {
  isAuthenticated: boolean;
  userEmail: string | undefined;
  isAuthPage: boolean;
  onLogout: () => void;
}) {
  return (
    <div className="flex items-center gap-3">
      {isAuthenticated && userEmail && (
        <span className="hidden max-w-[160px] truncate text-sm text-slate-500 sm:block">
          {userEmail}
        </span>
      )}
      {isAuthenticated ? (
        <button
          onClick={onLogout}
          className="rounded-lg border border-slate-700 px-3 py-1.5 text-sm font-medium text-slate-400 hover:text-slate-200 cursor-pointer"
        >
          Sign Out
        </button>
      ) : !isAuthPage ? (
        <Link
          to="/login"
          className="rounded-lg bg-sky-400 px-4 py-2 text-sm font-semibold text-slate-950 hover:bg-sky-300"
        >
          Sign In
        </Link>
      ) : null}
    </div>
  );
}

export function ErrorBoundary({ error }: Route.ErrorBoundaryProps) {
  let title = "Something went wrong";
  let message = "An unexpected error occurred.";
  let stack: string | undefined;

  if (isRouteErrorResponse(error)) {
    title = error.status === 404 ? "404" : `${error.status}`;
    message =
      error.status === 404
        ? "The page you are looking for does not exist."
        : error.statusText || message;
  } else if (error instanceof Error) {
    message = error.message;
    if (import.meta.env.DEV) {
      stack = error.stack;
    }
  }

  return (
    <html lang="en">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <Meta />
        <Links />
      </head>
      <body>
        <div className="flex min-h-screen items-center justify-center bg-slate-950 px-6">
          <div className="w-full max-w-lg rounded-xl border border-slate-800 bg-slate-900 p-8">
            <h1 className="mb-2 text-3xl font-bold text-slate-100">{title}</h1>
            <p className="mb-6 text-slate-400">{message}</p>
            <Link
              to="/"
              className="inline-flex rounded-lg bg-sky-400 px-4 py-2 font-medium text-slate-950 hover:bg-sky-300"
            >
              Return Home
            </Link>
            {stack && (
              <pre className="mt-6 overflow-x-auto rounded-lg border border-slate-800 bg-slate-950 p-4 text-xs text-red-400">
                <code>{stack}</code>
              </pre>
            )}
          </div>
        </div>
      </body>
    </html>
  );
}
