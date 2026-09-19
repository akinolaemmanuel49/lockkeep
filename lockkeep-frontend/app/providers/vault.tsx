import {
  createContext,
  useContext,
  useState,
  useRef,
  useCallback,
  useEffect,
  type ReactNode,
} from "react";
import { type VaultItem, type UpdateVaultItemRequest, type KDFParams, type CryptoPolicy } from "~/types/index";
import {
  deriveKeys,
  encryptSecret,
  decryptSecret,
  generateSalt,
  buildKDFParams,
} from "~/lib/crypto";

import { useAuth } from "./auth";
import {
  createVaultItem,
  deleteVaultItem,
  fetchVaultItems,
  updateVaultItem,
  updateVaultPassword,
  verifyVaultPassword,
} from "~/lib/api/vault";
import { fetchKDFParams } from "~/lib/api/auth";
import { getCurrentPolicy } from "~/lib/api/crypto.policy";
import { isSameKdfPolicy } from "~/utils/isSameKindKdfPolicy";

const VAULT_EXPIRY_MS = 5 * 60 * 1000;

interface VaultToken {
  key: CryptoKey;
  createdAt: number;
}

interface VaultContextValue {
  isLocked: boolean;
  items: VaultItem[];
  isLoading: boolean;
  unlockVault: (password: string) => Promise<void>;
  lockVault: () => void;
  getSecret: (itemId: string) => Promise<string>;
  addItem: (data: {
    type: "login";
    name: string;
    metadata: {
      siteUrl?: string;
      identifier?: string;
      notes?: string;
    };
    secret: string;
  }) => Promise<void>;
  changeVaultPassword: (
    oldPassword: string,
    newPassword: string,
  ) => Promise<void>;
  updateItem: (
    id: string,
    updates: Partial<{
      type: "login";
      name: string;
      metadata: {
        siteUrl?: string;
        identifier?: string;
        notes?: string;
      };
      secret: string;
    }>,
  ) => Promise<void>;
  deleteItem: (id: string) => Promise<void>;
}

const VaultContext = createContext<VaultContextValue | null>(null);

export function VaultProvider({ children }: { children: ReactNode }) {
  const [items, setItems] = useState<VaultItem[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isLocked, setIsLocked] = useState(true);
  const { getAccessToken, isAuthenticated, user } = useAuth();
  const [updateKdf, setUpdateKdf] = useState(false);

  const tokenRef = useRef<VaultToken | null>(null);
  const secretCache = useRef<Map<string, string>>(new Map());
  const lastUserIdRef = useRef<string | null>(null);
  const currentPolicy = useRef<CryptoPolicy>(null);
  const currentKDFParams = useRef<Omit<KDFParams, 'salt'>>(null);
  const userKdf = user?.vault?.kdf;

  useEffect(() => {
    const currentUserId = user?.id ?? null;

    // If user changed (including logout → null), wipe everything
    if (currentUserId !== lastUserIdRef.current) {
      tokenRef.current = null;
      secretCache.current.clear();
      setItems([]);
      setIsLocked(true);
      lastUserIdRef.current = currentUserId;
    }
  }, [user?.id]);

  useEffect(() => {
    const loadItems = async () => {
      if (!getAccessToken()) {
        setItems([]);
        return;
      }

      setIsLoading(true);
      try {
        setItems(await fetchVaultItems());
      } catch (err) {
        console.error("Failed to load vault items:", err);
      } finally {
        setIsLoading(false);
      }
    };

    loadItems();
  }, [getAccessToken, isAuthenticated]);

  useEffect(() => {
    const loadCurrentPolicy = async () => {
      if (!getAccessToken() || !user) {
        return;
      }

      currentPolicy.current = await getCurrentPolicy();
      currentKDFParams.current = currentPolicy.current.kdfParams;

      const needsKdfUpgrade =
        userKdf && !isSameKdfPolicy(currentKDFParams.current!, userKdf);

      setUpdateKdf(!!needsKdfUpgrade);
    };

    loadCurrentPolicy();
  }, [getAccessToken, isAuthenticated, user?.id]);

  const isExpired = useCallback((): boolean => {
    const token = tokenRef.current;
    if (!token) return true;
    return Date.now() - token.createdAt > VAULT_EXPIRY_MS;
  }, []);

  const checkLocked = useCallback((): boolean => {
    const expired = isExpired();
    if (expired && tokenRef.current) {
      tokenRef.current = null;
      secretCache.current.clear();
      setIsLocked(true);
    }
    return expired;
  }, [isExpired]);

  const getAccessTokenOrThrow = useCallback((): string => {
    const token = getAccessToken();
    if (!token) throw new Error("Not authenticated");
    return token;
  }, [getAccessToken]);

  /**
   * Deduplicated core: re-encrypt all vault items with a new key,
   * update server state, and refresh local cache.
   */
  const reencryptAllItems = useCallback(
    async (
      oldKey: CryptoKey,
      newKey: CryptoKey,
      newVerificationHash: string,
      newKdfParams: KDFParams,
      sourceItems: VaultItem[],
    ): Promise<VaultItem[]> => {
      const updatedItems: VaultItem[] = [];

      for (const item of sourceItems) {
        let plaintext = secretCache.current.get(item.id);
        if (!plaintext) {
          plaintext = await decryptSecret({
            ciphertext: item.secret.ciphertext,
            iv: item.secret.iv,
            tag: item.secret.tag,
            version: item.secret.version!,
          }, oldKey);
        }

        const newSecret = await encryptSecret(plaintext, newKey);
        const markForUpdate = currentPolicy.current?.version !== user?.vault?.version;

        const updated = await updateVaultItem(item.id, {
          type: item.type,
          name: item.name,
          metadata: item.metadata || {},
          secret: newSecret,
        }, markForUpdate);

        updatedItems.push(updated);
        secretCache.current.set(item.id, plaintext);
      }

      await updateVaultPassword(newVerificationHash, newKdfParams);

      return updatedItems;
    },
    [],
  );

  const unlockVault = useCallback(
    async (password: string) => {
      getAccessTokenOrThrow();
      setIsLoading(true);

      try {
        const kdfParams = await fetchKDFParams();
        const { encryptionKey, verificationHash } = await deriveKeys(
          password,
          kdfParams,
        );

        const result = await verifyVaultPassword(verificationHash);
        if (!result.success) {
          throw new Error("Invalid vault password");
        }

        const cache = new Map<string, string>();
        for (const item of result.credentials) {
          cache.set(item.id, await decryptSecret({
            ciphertext: item.secret.ciphertext,
            iv: item.secret.iv,
            tag: item.secret.tag,
            version: item.secret.version!,
          }, encryptionKey));
        }

        let finalItems = result.credentials;
        let finalKey = encryptionKey;

        // KDF upgrade path: same password, new KDF params
        if (updateKdf && currentKDFParams.current) {
          const newSalt = generateSalt();
          const newKdfParams = buildKDFParams(currentKDFParams.current as KDFParams, newSalt);
          const { encryptionKey: upgradedKey, verificationHash: upgradedHash } =
            await deriveKeys(password, newKdfParams);

          finalItems = await reencryptAllItems(
            encryptionKey,
            upgradedKey,
            upgradedHash,
            newKdfParams,
            result.credentials,
          );

          finalKey = upgradedKey;
          setUpdateKdf(false);
        }

        tokenRef.current = { key: finalKey, createdAt: Date.now() };
        secretCache.current = cache;
        setItems(finalItems);
        setIsLocked(false);
      } finally {
        setIsLoading(false);
      }
    },
    [getAccessTokenOrThrow, updateKdf, reencryptAllItems],
  );

  const lockVault = useCallback(() => {
    tokenRef.current = null;
    secretCache.current.clear();
    setIsLocked(true);
  }, []);

  const getSecret = useCallback(
    async (itemId: string): Promise<string> => {
      if (checkLocked()) throw new Error("Vault locked");

      const cached = secretCache.current.get(itemId);
      if (cached) return cached;

      const item = items.find((i) => i.id === itemId);
      if (!item) throw new Error("Vault item not found");

      const decrypted = await decryptSecret({
        ciphertext: item.secret.ciphertext,
        iv: item.secret.iv,
        tag: item.secret.tag,
        version: item.secret.version!,
      }, tokenRef.current!.key);
      secretCache.current.set(itemId, decrypted);
      return decrypted;
    },
    [checkLocked, items],
  );

  const addItem = useCallback(
    async (data: {
      type: "login";
      name: string;
      metadata: {
        siteUrl?: string;
        identifier?: string;
        notes?: string;
      };
      secret: string;
    }) => {
      getAccessTokenOrThrow();
      if (checkLocked()) throw new Error("Vault locked");

      setIsLoading(true);
      try {
        const secret = await encryptSecret(data.secret, tokenRef.current!.key);

        const newItem = await createVaultItem({
          type: data.type,
          name: data.name,
          metadata: data.metadata,
          secret,
        });

        secretCache.current.set(newItem.id, data.secret);
        setItems((prev) => [newItem, ...prev]);
      } finally {
        setIsLoading(false);
      }
    },
    [getAccessTokenOrThrow, checkLocked],
  );

  const changeVaultPassword = useCallback(
    async (oldPassword: string, newPassword: string) => {
      getAccessTokenOrThrow();

      const currentKdf = await fetchKDFParams();
      const { encryptionKey: oldKey, verificationHash: oldHash } =
        await deriveKeys(oldPassword, currentKdf);

      const verifyResult = await verifyVaultPassword(oldHash);
      if (!verifyResult.success) {
        throw new Error("Invalid current password");
      }

      const newSalt = generateSalt();
      const newKdfParams = buildKDFParams(currentKdf, newSalt);
      const { encryptionKey: newKey, verificationHash: newVerificationHash } =
        await deriveKeys(newPassword, newKdfParams);

      const updatedItems = await reencryptAllItems(
        oldKey,
        newKey,
        newVerificationHash,
        newKdfParams,
        items,
      );

      tokenRef.current = { key: newKey, createdAt: Date.now() };
      setItems(updatedItems);
      setUpdateKdf(false);
    },
    [getAccessTokenOrThrow, items, reencryptAllItems],
  );

  const updateItem = useCallback(
    async (
      id: string,
      updates: Partial<{
        type: "login";
        name: string;
        metadata: {
          siteUrl?: string;
          identifier?: string;
          notes?: string;
        };
        secret: string;
      }>,
    ) => {
      getAccessTokenOrThrow();
      if (checkLocked()) throw new Error("Vault locked");

      setIsLoading(true);
      try {
        const item = items.find((i) => i.id === id);
        if (!item) throw new Error("Vault item not found");

        const serverUpdates: UpdateVaultItemRequest = {
          type: updates.type || item.type,
          name: updates.name || item.name,
          metadata: { ...(item.metadata || {}), ...(updates.metadata || {}) },
          secret: item.secret,
        };

        if (updates.secret !== undefined) {
          serverUpdates.secret = await encryptSecret(
            updates.secret,
            tokenRef.current!.key,
          );
        }

        const updated = await updateVaultItem(id, serverUpdates);
        setItems((prev) => prev.map((i) => (i.id === id ? updated : i)));

        if (updates.secret !== undefined) {
          secretCache.current.set(id, updates.secret);
        }
      } finally {
        setIsLoading(false);
      }
    },
    [getAccessTokenOrThrow, checkLocked, items],
  );

  const deleteItem = useCallback(
    async (id: string) => {
      getAccessTokenOrThrow();
      setIsLoading(true);
      try {
        await deleteVaultItem(id);
        secretCache.current.delete(id);
        setItems((prev) => prev.filter((i) => i.id !== id));
      } finally {
        setIsLoading(false);
      }
    },
    [getAccessTokenOrThrow],
  );

  return (
    <VaultContext.Provider
      value={{
        isLocked,
        items,
        isLoading,
        unlockVault,
        lockVault,
        getSecret,
        addItem,
        changeVaultPassword,
        updateItem,
        deleteItem,
      }}
    >
      {children}
    </VaultContext.Provider>
  );
}

export function useVault(): VaultContextValue {
  const ctx = useContext(VaultContext);
  if (!ctx) throw new Error("useVault must be used within VaultProvider");
  return ctx;
}