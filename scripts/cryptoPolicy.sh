// seed-crypto-policy.js

use("lockkeep");

// Remove any existing policy
db.cryptopolicy.deleteMany({});

// -----------------------------------------------------------------------------
// Current Crypto Policy (v1 - Scrypt)
// -----------------------------------------------------------------------------
db.cryptopolicy.insertOne({
    _id: "current",

    version: 1,

    kdf_params: {
        algorithm: "scrypt",

        // scrypt parameters
        // N = 2^17
        memory: 128,
        iterations: 17,
        parallelism: 1
    },

    updated_at: new Date()
});


// -----------------------------------------------------------------------------
// Migration Target (v2 - Argon2id)
// Uncomment this block and comment out the one above to test migration.
// -----------------------------------------------------------------------------

/*
db.cryptopolicy.insertOne({
    _id: "current",

    version: 2,

    kdf_params: {
        algorithm: "argon2id",

        // Argon2id parameters
        memory: 65536,      // 64 MiB
        iterations: 3,
        parallelism: 1
    },

    updated_at: new Date()
});
*/

print("Crypto policy seeded successfully.");