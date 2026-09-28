import { Link } from "react-router";
import type { Route } from "./+types/not-found";
import VaultDial from "~/components/VaultDial";
import Nameplate from "~/components/Nameplate";

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
    <div className="flex min-h-[calc(100vh-9rem)] flex-col items-center justify-center px-4 text-center">
      <VaultDial size={160} pointer="?" className="text-brass-600" />

      <Nameplate className="mt-8 justify-center" tone="muted">
        Error 404
      </Nameplate>
      <h1 className="mt-3 font-display text-4xl font-semibold tracking-tight text-ivory">
        Off the record.
      </h1>
      <p className="mt-3 mb-8 max-w-md text-sand-400">
        This case may not be part of the vault's index. The combination you
        dialed doesn't correspond to anything on file.
      </p>

      <Link to="/" className="lk-btn lk-btn--primary">
        Return to the Vault
      </Link>
    </div>
  );
}