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
import BrandLock from "./components/BrandLock";
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
    href: "https://fonts.googleapis.com/css2?family=Fraunces:ital,opsz,wght@0,9..144,300..900;1,9..144,300..900&family=JetBrains+Mono:wght@400;500;700&family=Space+Grotesk:wght@300..700&display=swap",
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
        <div className="lk-ambient flex min-h-screen flex-col font-sans text-ivory antialiased">
          <header className="sticky top-0 z-50 border-b border-brass-500/20 bg-coal-950/85 backdrop-blur">
            <div className="mx-auto flex h-16 max-w-6xl items-center justify-between gap-4 px-6">
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

          <main className="mx-auto w-full max-w-6xl flex-1 px-6 py-10">
            <Outlet />
          </main>

          <footer className="border-t border-brass-500/10 py-5">
            <div className="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-2 px-6 text-[0.6875rem] uppercase tracking-[0.3em] text-sand-600">
              <span>LockKeep Vault</span>
              <span className="font-mono tracking-[0.22em]">
                {/* design-time distinction: v2 workspace-aware UI */}
                v2 :: workspace-aware
              </span>
            </div>
          </footer>

          <ToastContainer />
        </div>
      </body>
    </html>
  );
}

function Logo() {
  return (
    <Link to="/" className="group flex items-center gap-3">
      <BrandLock className="h-8 w-8 text-brass-500 transition-colors group-hover:text-brass-400" />
      <span className="flex flex-col leading-none">
        <span className="font-display text-xl font-semibold tracking-tight text-ivory">
          LockKeep
        </span>
        <span className="mt-0.5 text-[0.5625rem] font-semibold uppercase tracking-[0.32em] text-brass-500/80">
          The Vault
        </span>
      </span>
    </Link>
  );
}

function Nav() {
  const { user } = useAuth();
  const { isLocked } = useVault();

  return (
    <nav className="hidden items-center gap-1 md:flex">
      <NavLink to="/dashboard">
        Vault
        <span
          className={`ml-2 inline-block h-1.5 w-1.5 rounded-full ${
            isLocked ? "bg-vermillion-400" : "bg-patina-400"
          }`}
        />
      </NavLink>
      <NavLink to="/workspaces" matchPrefix>
        Workspaces
      </NavLink>
      <NavLink to="/settings">Settings</NavLink>
      {user?.systemRole === "system:admin" && (
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
        <span className="hidden max-w-[180px] truncate font-mono text-xs tracking-wide text-sand-500 lg:block">
          {userEmail}
        </span>
      )}
      {isAuthenticated ? (
        <button onClick={onLogout} className="lk-btn lk-btn--ghost">
          Sign Out
        </button>
      ) : !isAuthPage ? (
        <Link to="/login" className="lk-btn lk-btn--primary">
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
        <div className="lk-ambient flex min-h-screen items-center justify-center px-6">
          <div className="plate w-full max-w-lg p-8">
            <h1 className="mb-2 font-display text-3xl font-semibold text-ivory">
              {title}
            </h1>
            <p className="mb-6 text-sm text-sand-400">{message}</p>
            <Link to="/" className="lk-btn lk-btn--primary">
              Return to the Vault
            </Link>
            {stack && (
              <pre className="mt-6 overflow-x-auto rounded-lg border border-vermillion-500/30 bg-coal-950 p-4 text-xs text-vermillion-300">
                <code>{stack}</code>
              </pre>
            )}
          </div>
        </div>
      </body>
    </html>
  );
}