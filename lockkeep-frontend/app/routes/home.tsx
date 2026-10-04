import type { Route } from "./+types/home";
import { Link } from "react-router";

import { requireGuest } from "~/lib/auth-guard";
import VaultDial from "~/components/VaultDial";
import Nameplate from "~/components/Nameplate";

export const clientLoader = () => {
  return requireGuest();
};

export function meta({}: Route.MetaArgs) {
  return [
    { title: "LockKeep — Zero-Trust Password Vault" },
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
    <div className="flex flex-col items-center justify-center gap-14 py-12 text-center md:flex-row md:gap-20 md:py-20 md:text-left">
      {/* Dial */}
      <div className="relative shrink-0">
        <div
          aria-hidden="true"
          className="absolute -inset-10 rounded-full bg-[radial-gradient(closest-side,rgba(201,162,39,0.16),transparent)]"
        />
        <VaultDial size={224} spinning pointer="0" className="relative" />
      </div>

      {/* Copy */}
      <div className="max-w-xl">
        <Nameplate className="justify-center md:justify-start">
          Mission Control
        </Nameplate>

        <h1 className="mt-4 font-display text-4xl font-semibold leading-tight tracking-tight text-ivory sm:text-5xl">
          Your secrets, <span className="text-brass-sheen">locked in.</span>
        </h1>

        <p className="mt-5 max-w-md text-base leading-relaxed text-sand-400">
          Every credential is encrypted on your device before it ever crosses
          the wire. We cannot see, access, or decrypt your passwords. That's the
          point of the vault.
        </p>

        <div className="mt-8 flex flex-wrap items-center justify-center gap-4 md:justify-start">
          <Link to="/signup" className="lk-btn lk-btn--primary px-7 py-3 text-sm">
            Open Your Vault
          </Link>
          <Link to="/login" className="lk-btn lk-btn--ghost px-7 py-3 text-sm">
            Sign In
          </Link>
        </div>
      </div>
    </div>
  );
}