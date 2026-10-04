db.users.insertOne(
    {
    _id: ObjectId('6a48d169d26388ac5f48fc2b'),
    username: '',
    email: 'admin@example.com',
    tenant_id: 'tenant_6a48d169',
    user_type: 'admin',
    auth_method: 'local',
    auth_provider_id: null,
    password_hash: '$2a$10$X4TvamAeIyMCS3uMY8Hk/uDfnjj2t/.jHqSGigCbLSkpKw4KNeT2K',
    vault: {
        version: 1,
        verification_hash: '029044da3786584e20aa41926e34973fd916aacb3e5dd71a0fcf3ce3fd648f90',
        kdf: {
            algorithm: 'scrypt',
            salt: '8ce036c8ced6fef7ecbb045101764e59179e644873e42c07096a9cc51c6d457b',
            memory: 128,
            iterations: 17,
            parallelism: 1
        }
    },
    created_at: ISODate('2026-07-04T09:24:57.383Z'),
    updated_at: ISODate('2026-07-04T09:25:09.451Z')
});