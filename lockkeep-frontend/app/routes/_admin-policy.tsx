import { useState } from "react";
import type { KDFAlgorithm } from "~/lib/crypto";
import { useAuth } from "~/providers/auth";
import { useToast } from "~/providers/toast";
import type { Route } from "./+types/_admin-policy";
import { setCurrentPolicy } from "~/lib/api/crypto.policy";

export function meta({ }: Route.MetaArgs) {
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
    const [form, setForm] = useState<PolicyFormData>({
        algorithm: "argon2id",
        memory: 65536,
        iterations: 3,
        parallelism: 4,
    });

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
        }

        try {
            await setCurrentPolicy(reqBody)

            addToast("Crypto policy updated successfully", "success");

        } catch (err: any) {
            addToast(err.message, "error");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="mx-auto max-w-2xl">
            <h1 className="mb-6 text-2xl font-bold text-slate-100">
                Cryptographic Policy
            </h1>

            <div className="rounded-xl border border-slate-800 bg-slate-900 p-6">
                <form onSubmit={handleSubmit} className="space-y-6">
                    <div>
                        <label className="mb-2 block text-sm font-medium text-slate-400">
                            KDF Algorithm
                        </label>
                        <select
                            value={form.algorithm}
                            onChange={(e) => setForm({ ...form, algorithm: e.target.value as KDFAlgorithm })}
                            className="w-full rounded-lg border border-slate-700 bg-slate-950 px-4 py-2.5 text-slate-200 focus:border-sky-400 focus:outline-none"
                        >
                            <option value="argon2id">Argon2id</option>
                            <option value="scrypt">Scrypt</option>
                        </select>
                    </div>

                    <div className="grid gap-6 sm:grid-cols-2">
                        <div>
                            <label className="mb-2 block text-sm font-medium text-slate-400">
                                Memory ({form.algorithm === "scrypt" ? "MB" : "KB"})
                            </label>
                            <input
                                type="number"
                                value={form.memory}
                                onChange={(e) => setForm({ ...form, memory: Number(e.target.value) })}
                                className="w-full rounded-lg border border-slate-700 bg-slate-950 px-4 py-2.5 text-slate-200 focus:border-sky-400 focus:outline-none"
                            />
                        </div>
                        <div>
                            <label className="mb-2 block text-sm font-medium text-slate-400">
                                Iterations
                            </label>
                            <input
                                type="number"
                                value={form.iterations}
                                onChange={(e) => setForm({ ...form, iterations: Number(e.target.value) })}
                                className="w-full rounded-lg border border-slate-700 bg-slate-950 px-4 py-2.5 text-slate-200 focus:border-sky-400 focus:outline-none"
                            />
                        </div>
                    </div>

                    <div>
                        <label className="mb-2 block text-sm font-medium text-slate-400">
                            Parallelism
                        </label>
                        <input
                            type="number"
                            min={1}
                            max={8}
                            value={form.parallelism}
                            onChange={(e) => setForm({ ...form, parallelism: Number(e.target.value) })}
                            className="w-full rounded-lg border border-slate-700 bg-slate-950 px-4 py-2.5 text-slate-200 focus:border-sky-400 focus:outline-none"
                        />
                    </div>

                    <div className="flex items-center gap-4 pt-2">
                        <button
                            type="submit"
                            disabled={loading}
                            className="rounded-lg bg-sky-400 px-6 py-2.5 text-sm font-semibold text-slate-950 hover:bg-sky-300 disabled:opacity-50"
                        >
                            {loading ? "Updating..." : "Update Policy"}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
}