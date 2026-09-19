import type React from "react";
import { Link, useLocation } from "react-router";

export default function NavLink({
  to,
  children,
}: {
  to: string;
  children: React.ReactNode;
}) {
  const location = useLocation();
  const isActive = location.pathname === to;

  return (
    <Link
      to={to}
      className={`rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
        isActive
          ? "bg-slate-800 text-slate-100"
          : "text-slate-400 hover:text-slate-200"
      }`}
    >
      {children}
    </Link>
  );
}
