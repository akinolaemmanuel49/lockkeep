import type { Route } from "./+types/home";
import { Link } from "react-router";

import { requireGuest } from "~/lib/auth-guard";

export const clientLoader = () => {
  return requireGuest();
};

export function meta({}: Route.MetaArgs) {
  return [
    { title: "LockKeep - Zero-Trust Password Vault" },
    {
      name: "description",
      content:
        "Secure, client-side encrypted password manager. We cannot see, access, or decrypt your passwords. Ever.",
    },
    {
      name: "keywords",
      content: "password manager, zero trust, encryption, secure vault",
    },
    { property: "og:title", content: "LockKeep - Zero-Trust Password Vault" },
    {
      property: "og:description",
      content:
        "Secure, client-side encrypted password manager. We cannot see, access, or decrypt your passwords.",
    },
    { property: "og:type", content: "website" },
  ];
}

export default function Home() {
  return (
    <div className="flex min-h-[calc(100vh-8rem)] flex-col items-center justify-center px-4 text-center">
      <div className="mb-6 text-sky-400">
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
          <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
          <path d="M7 11V7a5 5 0 0 1 10 0v4" />
        </svg>
      </div>

      <h1 className="mb-4 text-4xl font-bold tracking-tight text-slate-100">
        Zero-Trust Password Vault
      </h1>

      <p className="mb-10 max-w-lg text-lg text-slate-400">
        Your credentials are encrypted on your device before they ever reach our
        servers. We cannot see, access, or decrypt your passwords. Ever.
      </p>

      <div className="mb-10 flex flex-wrap justify-center gap-8">
        <Feature
          icon="shield"
          text="Client-side encryption"
          color="text-green-400"
        />
        <Feature
          icon="lock"
          text="Vault password protected"
          color="text-amber-400"
        />
        <Feature
          icon="users"
          text="Multi-tenant isolation"
          color="text-violet-400"
        />
      </div>

      <div className="flex gap-4">
        <Link
          to="/signup"
          className="rounded-xl bg-sky-400 px-8 py-3.5 text-base font-semibold text-slate-950 hover:bg-sky-300"
        >
          Get Started
        </Link>
        <Link
          to="/login"
          className="rounded-xl border border-slate-700 px-8 py-3.5 text-base font-medium text-slate-200 hover:border-slate-600"
        >
          Sign In
        </Link>
      </div>
    </div>
  );
}

function Feature({
  icon,
  text,
  color,
}: {
  icon: "shield" | "lock" | "users";
  text: string;
  color: string;
}) {
  const paths = {
    shield: <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />,
    lock: (
      <>
        <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
        <path d="M7 11V7a5 5 0 0 1 10 0v4" />
      </>
    ),
    users: (
      <>
        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
        <circle cx="9" cy="7" r="4" />
        <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
        <path d="M16 3.13a4 4 0 0 1 0 7.75" />
      </>
    ),
  };

  return (
    <div className="flex items-center gap-2.5">
      <svg
        width="20"
        height="20"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        className={color}
      >
        {paths[icon]}
      </svg>
      <span className="text-sm font-medium text-slate-300">{text}</span>
    </div>
  );
}
