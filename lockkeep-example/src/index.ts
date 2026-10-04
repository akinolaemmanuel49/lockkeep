#!/usr/bin/env node
import { spawn } from "node:child_process";
import {
    fetchEnvVars,
    fetchSecrets,
    LockKeepError,
    type DecryptedSecret,
} from "@lockkeep/sdk";

const apiKey = (process.env.LOCKKEEP_API_KEY ?? "").trim();
if (!apiKey) {
    console.error(
        "LOCKKEEP_API_KEY is required. Mint an environment-scoped key in the" +
            " LockKeep console (Service accounts) and set it in your environment.",
    );
    process.exit(1);
}

const baseUrl =
    (process.env.LOCKKEEP_API_URI ?? "").trim() ||
    "http://localhost:8000/api/v2";
const args = process.argv.slice(2);
const showValues = args.includes("--env");
const assert = args.includes("--assert");

const byKey = (a: DecryptedSecret, b: DecryptedSecret) =>
    a.key < b.key ? -1 : a.key > b.key ? 1 : 0;

async function printSecrets(): Promise<void> {
    const set = await fetchSecrets({ apiKey, baseUrl });
    const secrets = [...set.secrets].sort(byKey);

    console.log(
        `LockKeep env ${set.environmentId}: ${secrets.length} secret(s) fetched in a single API call`,
    );

    if (showValues) {
        for (const s of secrets) console.log(`${s.key}=${s.value}`);
    } else {
        for (const s of secrets) {
            console.log(`${s.key}=<hidden> (${s.kind}, v${s.version})`);
        }
        console.log("Run with --env to print values.");
    }

    if (assert && secrets.length === 0) {
        console.error(`No secrets returned for this key (env ${set.environmentId}).`);
        process.exit(3);
    }
}

function runChild(
    command: string,
    commandArgs: string[],
    vars: Record<string, string>,
): Promise<number> {
    return new Promise((resolve, reject) => {
        const child = spawn(command, commandArgs, {
            stdio: "inherit",
            env: { ...process.env, ...vars },
        });
        child.on("error", reject);
        child.on("exit", (code, signal) => {
            if (signal !== null) {
                console.error(`child terminated by signal ${signal}`);
                resolve(1);
                return;
            }
            resolve(code ?? 1);
        });
    });
}

async function run(): Promise<void> {
    if (args[0] === "run") {
        const cmdArgs = args.slice(1).filter((a) => a !== "--");
        if (cmdArgs.length === 0) {
            console.error("Usage: lockkeep-example run -- <command> [args...]");
            process.exit(2);
        }

        // One API call, then hand the secrets over as plain environment
        // variables — the app never needs to know about LockKeep.
        const vars = await fetchEnvVars({ apiKey, baseUrl });
        const set = Object.keys(vars);
        console.log(
            `LockKeep env injected: ${set.length} variable(s) available as process.env`,
        );

        const code = await runChild(cmdArgs[0]!, cmdArgs.slice(1), vars);
        if (code !== 0) process.exit(code);
        return;
    }

    await printSecrets();
}

run().catch((err) => {
    console.error(err instanceof LockKeepError ? `[lockkeep] ${err.message}` : err);
    process.exit(1);
});