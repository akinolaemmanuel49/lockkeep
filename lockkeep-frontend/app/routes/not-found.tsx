import { Link } from "react-router";
import type { Route } from "./+types/not-found";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Page Not Found - LockKeep" },
    {
      name: "description",
      content: "The page you are looking for does not exist.",
    },
    { name: "robots", content: "noindex, nofollow" },
  ];
}

export default function NotFound() {
  return (
    <div className="flex min-h-[calc(100vh-8rem)] flex-col items-center justify-center px-4 text-center">
      <div className="mb-6 text-slate-600">
        <svg
          width="64"
          height="64"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <circle cx="12" cy="12" r="10" />
          <line x1="12" y1="8" x2="12" y2="12" />
          <line x1="12" y1="16" x2="12.01" y2="16" />
        </svg>
      </div>

      <h1 className="mb-4 text-4xl font-bold tracking-tight text-slate-100">
        404
      </h1>

      <p className="mb-10 max-w-lg text-lg text-slate-400">
        The page you are looking for does not exist.
      </p>

      <Link
        to="/"
        className="rounded-xl bg-sky-400 px-8 py-3.5 text-base font-semibold text-slate-950 hover:bg-sky-300"
      >
        Return Home
      </Link>
    </div>
  );
}
