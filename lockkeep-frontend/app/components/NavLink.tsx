import type React from "react";
import { Link, useLocation } from "react-router";

export default function NavLink({
  to,
  children,
  matchPrefix = false,
}: {
  to: string;
  children: React.ReactNode;
  matchPrefix?: boolean;
}) {
  const location = useLocation();
  const isActive = matchPrefix
    ? location.pathname === to || location.pathname.startsWith(`${to}/`)
    : location.pathname === to;

  return (
    <Link
      to={to}
      className={`group inline-flex items-center rounded-md px-3 py-2 text-[0.8125rem] font-medium uppercase tracking-[0.1em] transition-colors ${
        isActive
          ? "text-brass-300"
          : "text-sand-400 hover:text-ivory"
      }`}
    >
      <span
        aria-hidden="true"
        className={`mr-2 h-3.5 w-px bg-current transition-opacity ${
          isActive ? "opacity-80" : "opacity-20 group-hover:opacity-60"
        }`}
      />
      {children}
    </Link>
  );
}