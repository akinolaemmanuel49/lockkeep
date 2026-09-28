import { Link } from "react-router";
import Nameplate from "~/components/Nameplate";

export default function Unauthorized() {
  return (
    <div className="flex flex-col items-center justify-center py-24 text-center">
      <Nameplate tone="danger" className="justify-center">
        Access Denied
      </Nameplate>
      <h1 className="mt-4 font-display text-4xl font-semibold tracking-tight text-ivory">
        403
      </h1>
      <p className="mt-3 mb-8 max-w-sm text-sm text-sand-400">
        Your clearance does not permit access to this section. This attempt has
        not been logged... probably.
      </p>
      <Link to="/dashboard" className="lk-btn lk-btn--primary">
        Return to the Vault
      </Link>
    </div>
  );
}