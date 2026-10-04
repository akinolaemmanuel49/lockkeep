#!/usr/bin/env node
// Example service that reads injected secrets as plain environment variables.
// It never imports @lockkeep/sdk — lockkeep-example run injected them.
const showValues = process.argv.includes("--values");

for (const key of ["DATABASE_URL", "WRITEUP"]) {
    const value = process.env[key];
    if (value === undefined) {
        console.log(`${key}=<unset>`);
    } else if (showValues) {
        console.log(`${key}=${value}`);
    } else {
        console.log(`${key}=<set>`);
    }
}

if (!process.env.DATABASE_URL || !process.env.WRITEUP) {
    console.error("secrets are missing from the environment");
    process.exit(1);
}

console.log("demo app ready — secrets were injected as environment variables");