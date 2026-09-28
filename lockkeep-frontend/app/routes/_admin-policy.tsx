import { useEffect, useState } from "react";
import type { KDFAlgorithm } from "~/lib/crypto";
import { useToast } from "~/providers/toast";
import type { Route } from "./+types/_admin-policy";
import { getCurrentPolicy, setCurrentPolicy } from "~/lib/api/crypto.policy";
import Nameplate from "~/components/Nameplate";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Crypto Policy — LockKeep Admin" },
    { name: "description", content: "Manage system-wide cryptographic policy" },
  ];
}

interface PolicyFormData {
  algorithm: KDFAlgorithm;
  memory: number;
  iterations: number;
  parallelism: number;
}

export default function AdminPolicy() {
  const { addToast } = useToast();
  const [loading, setLoading] = useState(false);
  const [isLoadingPolicy, setIsLoadingPolicy] = useState(true);
  const [currentVersion, setCurrentVersion] = useState<number | null>(null);
  const [updatedAt, setUpdatedAt] = useState<string | null>(null);
  const [form, setForm] = useState<PolicyFormData>({
    algorithm: "argon2id",
    memory: 65536,
    iterations: 3,
    parallelism: 4,
  });

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const policy = await getCurrentPolicy();
        if (cancelled) return;
        setForm({
          algorithm: policy.kdfParams.algorithm,
          memory: policy.kdfParams.memory,
          iterations: policy.kdfParams.iterations,
          parallelism: policy.kdfParams.parallelism,
        });
        setCurrentVersion(policy.version);
        setUpdatedAt(policy.updatedAt);
      } catch {
        // No policy served yet — keep the recommended defaults.
      } finally {
        if (!cancelled) setIsLoadingPolicy(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    const reqBody = {
      kdfParams: {
        algorithm: form.algorithm,
        memory: form.memory,
        iterations: form.iterations,
        parallelism: form.parallelism,
      },
    };

    try {
      await setCurrentPolicy(reqBody);
      addToast("Crypto policy updated successfully", "success");
      const latest = await getCurrentPolicy();
      setCurrentVersion(latest.version);
      setUpdatedAt(latest.updatedAt);
    } catch (err: any) {
      addToast(err.message, "error");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="mx-auto max-w-2xl">
      <div className="mb-6">
        <Nameplate>System Administration</Nameplate>
        <h1 className="mt-2 font-display text-3xl font-semibold tracking-tight text-ivory">
          Cryptographic Policy
        </h1>
      </div>

      <div className="plate p-6">
        <div className="mb-6 flex flex-wrap items-center justify-between gap-3 border-b border-brass-500/15 pb-4">
          <p className="text-xs leading-relaxed text-sand-500">
            {isLoadingPolicy
              ? "Reading current policy..."
              : currentVersion !== null
                ? "Parameters below are the live policy."
                : "No policy on file yet — saving will establish one (version 1)."}
          </p>
          {!isLoadingPolicy && currentVersion !== null && updatedAt && (
            <span className="font-mono text-[0.625rem] text-sand-600">
              v{currentVersion} · {new Date(updatedAt).toLocaleString()}
            </span>
          )}
        </div>

        <form onSubmit={handleSubmit} className="space-y-6">
          <div>
            <label className="mb-2 block text-xs font-semibold uppercase tracking-[0.14em] text-sand-500">
              KDF Algorithm
            </label>
            <select
              value={form.algorithm}
              onChange={(e) =>
                setForm({ ...form, algorithm: e.target.value as KDFAlgorithm })
              }
              className="lk-input"
            >
              <option value="argon2id">Argon2id</option>
              <option value="scrypt">Scrypt</option>
            </select>
          </div>

          <div className="grid gap-6 sm:grid-cols-2">
            <div>
              <label className="mb-2 block text-xs font-semibold uppercase tracking-[0.14em] text-sand-500">
                Memory ({form.algorithm === "scrypt" ? "MB" : "KB"})
              </label>
              <input
                type="number"
                value={form.memory}
                onChange={(e) =>
                  setForm({ ...form, memory: Number(e.target.value) })
                }
                className="lk-input"
              />
            </div>
            <div>
              <label className="mb-2 block text-xs font-semibold uppercase tracking-[0.14em] text-sand-500">
                Iterations
              </label>
              <input
                type="number"
                value={form.iterations}
                onChange={(e) =>
                  setForm({ ...form, iterations: Number(e.target.value) })
                }
                className="lk-input"
              />
            </div>
          </div>

          <div>
            <label className="mb-2 block text-xs font-semibold uppercase tracking-[0.14em] text-sand-500">
              Parallelism
            </label>
            <input
              type="number"
              min={1}
              max={8}
              value={form.parallelism}
              onChange={(e) =>
                setForm({ ...form, parallelism: Number(e.target.value) })
              }
              className="lk-input"
            />
          </div>

          <div className="flex items-center gap-4 pt-2">
            <button type="submit" disabled={loading} className="lk-btn lk-btn--primary">
              {loading ? "Updating..." : "Update Policy"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}