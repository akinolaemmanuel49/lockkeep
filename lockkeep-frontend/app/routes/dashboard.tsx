import { useState } from "react";
import { useVault } from "~/providers/vault";
import { useToast } from "~/providers/toast";
import VaultItemCard from "~/components/VaultItemCard";
import PasswordItemModal from "~/components/PasswordItemModal";
import ItemTypeSelector from "~/components/ItemTypeSelector";
import UnlockModal from "~/components/UnlockModal";
import VaultDial from "~/components/VaultDial";
import Nameplate from "~/components/Nameplate";
import type { VaultItem, VaultItemType } from "~/types/index";
import { requireAuth, getCurrentUser } from "~/lib/auth-guard";
import { redirect } from "react-router";
import type { Route } from "./+types/dashboard";

export const clientLoader = () => {
  const redirecting = requireAuth();
  if (redirecting) return redirecting;

  const user = getCurrentUser();
  if (user && !user.hasMasterPassword) {
    return redirect("/setup");
  }
  return null;
};

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Dashboard - LockKeep" },
    {
      name: "description",
      content:
        "Manage your encrypted vault items. Search, add, and organize your secrets securely.",
    },
    { name: "robots", content: "noindex, nofollow" },
  ];
}

export default function Dashboard() {
  const {
    isLocked,
    items,
    isLoading,
    addItem,
    updateItem,
    deleteItem,
    getSecret,
    lockVault,
  } = useVault();
  const { addToast } = useToast();
  const [searchQuery, setSearchQuery] = useState("");
  const [isTypeSelectorOpen, setIsTypeSelectorOpen] = useState(false);
  const [isPasswordModalOpen, setIsPasswordModalOpen] = useState(false);
  const [isUnlockModalOpen, setIsUnlockModalOpen] = useState(false);
  const [pendingAction, setPendingAction] = useState<(() => void) | null>(null);
  const [editingItem, setEditingItem] = useState<VaultItem | null>(null);
  const [editingSecret, setEditingSecret] = useState<string | null>(null);
  const [deleteConfirmId, setDeleteConfirmId] = useState<string | null>(null);

  const filtered = searchQuery.trim()
    ? items.filter(
        (item) =>
          item.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
          ((item.metadata?.siteUrl as string) || "")
            .toLowerCase()
            .includes(searchQuery.toLowerCase()) ||
          ((item.metadata?.identifier as string) || "")
            .toLowerCase()
            .includes(searchQuery.toLowerCase()),
      )
    : items;

  const requestUnlock = (action: () => void) => {
    if (!isLocked) {
      action();
      return;
    }
    setPendingAction(() => action);
    setIsUnlockModalOpen(true);
  };

  const handleUnlockSuccess = () => {
    setIsUnlockModalOpen(false);
    if (pendingAction) {
      pendingAction();
      setPendingAction(null);
    }
  };

  const handleAddClick = () => {
    setEditingItem(null);
    setEditingSecret(null);
    requestUnlock(() => setIsTypeSelectorOpen(true));
  };

  const handleTypeSelect = (type: VaultItemType) => {
    setIsTypeSelectorOpen(false);
    if (type === "login") {
      setIsPasswordModalOpen(true);
    }
  };

  const handleEdit = (item: VaultItem) => {
    setEditingItem(item);
    setEditingSecret(null);
    requestUnlock(async () => {
      try {
        const secret = await getSecret(item.id);
        setEditingSecret(secret);
        setIsPasswordModalOpen(true);
      } catch (err) {
        addToast(
          err instanceof Error ? err.message : "Failed to decrypt secret",
          "error",
        );
      }
    });
  };

  const handleSavePassword = async (data: {
    name: string;
    metadata: Record<string, unknown>;
    secret: string;
  }) => {
    try {
      if (editingItem) {
        await updateItem(editingItem.id, {
          type: "login",
          name: data.name,
          metadata: data.metadata,
          secret: data.secret,
        });
        addToast("Password updated", "success");
      } else {
        await addItem({
          type: "login",
          name: data.name,
          metadata: data.metadata,
          secret: data.secret,
        });
        addToast("Password saved", "success");
      }
      setIsPasswordModalOpen(false);
      setEditingItem(null);
      setEditingSecret(null);
    } catch (err) {
      addToast(err instanceof Error ? err.message : "Failed to save", "error");
    }
  };

  const handleDelete = (id: string) => {
    setDeleteConfirmId(id);
  };

  const confirmDelete = async () => {
    if (!deleteConfirmId) return;
    try {
      await deleteItem(deleteConfirmId);
      addToast("Item deleted", "success");
    } catch (err) {
      addToast(err instanceof Error ? err.message : "Failed to delete", "error");
    } finally {
      setDeleteConfirmId(null);
    }
  };

  const handleRevealSecret = (
    itemId: string,
    onReveal: (secret: string) => void,
  ) => {
    requestUnlock(async () => {
      try {
        const secret = await getSecret(itemId);
        onReveal(secret);
      } catch (err) {
        addToast(err instanceof Error ? err.message : "Failed to get secret", "error");
      }
    });
  };

  const handleCopySecret = (itemId: string) => {
    requestUnlock(async () => {
      try {
        const secret = await getSecret(itemId);
        await navigator.clipboard.writeText(secret);
        addToast("Password copied", "success");
      } catch (err) {
        addToast(err instanceof Error ? err.message : "Failed to copy", "error");
      }
    });
  };

  return (
    <div>
      {/* Page header */}
      <div className="mb-8 flex flex-wrap items-end justify-between gap-4">
        <div>
          <Nameplate>Personal Vault</Nameplate>
          <h1 className="mt-2 font-display text-3xl font-semibold tracking-tight text-ivory">
            Your Vault
          </h1>
        </div>
        <div className="flex items-center gap-3">
          <span className="inline-flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.2em]">
            <span
              className={`h-2 w-2 rounded-full ${
                isLocked ? "bg-vermillion-400 animate-lk-lamp" : "bg-patina-400"
              }`}
            />
            {isLocked ? "Locked" : "Unlocked"}
          </span>
          {!isLocked && (
            <button onClick={lockVault} className="lk-btn lk-btn--ghost">
              Lock Vault
            </button>
          )}
        </div>
      </div>

      {/* Toolbar */}
      <div className="mb-8 flex flex-wrap items-center gap-4">
        <div className="well relative flex-1 min-w-[280px] rounded-lg">
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            className="absolute left-3 top-1/2 -translate-y-1/2 text-sand-600"
          >
            <circle cx="11" cy="11" r="8" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search records by name, site, or identifier..."
            className="w-full bg-transparent py-2.5 pl-10 pr-9 text-sm text-ivory placeholder:text-sand-600 focus:outline-none"
          />
          {searchQuery && (
            <button
              onClick={() => setSearchQuery("")}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-sand-600 hover:text-sand-300"
              aria-label="Clear search"
            >
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <line x1="18" y1="6" x2="6" y2="18" />
                <line x1="6" y1="6" x2="18" y2="18" />
              </svg>
            </button>
          )}
        </div>

        <button onClick={handleAddClick} className="lk-btn lk-btn--primary px-5 py-2.5">
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2.2"
            strokeLinecap="round"
          >
            <line x1="12" y1="5" x2="12" y2="19" />
            <line x1="5" y1="12" x2="19" y2="12" />
          </svg>
          Add Record
        </button>
      </div>

      {/* Content */}
      {isLoading && items.length === 0 ? (
        <div className="flex items-center justify-center gap-3 py-24 text-sand-500">
          <div className="h-5 w-5 animate-spin rounded-full border-2 border-brass-600 border-t-brass-400" />
          <span>Cracking the vault open...</span>
        </div>
      ) : filtered.length === 0 ? (
        <div className="flex flex-col items-center py-20 text-center">
          <VaultDial size={120} className="text-brass-700 opacity-60" />
          <p className="mt-5 font-display text-lg font-medium text-sand-400">
            {searchQuery ? "No matching records" : "The vault is empty"}
          </p>
          <p className="mt-1 text-sm text-sand-600">
            {searchQuery
              ? "Try adjusting your search"
              : "Add a record to file your first secret."}
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-5 md:grid-cols-2 lg:grid-cols-3">
          {filtered.map((item) => (
            <VaultItemCard
              key={item.id}
              item={item}
              onRevealSecret={(onReveal) => handleRevealSecret(item.id, onReveal)}
              onCopySecret={() => handleCopySecret(item.id)}
              onEdit={handleEdit}
              onDelete={handleDelete}
            />
          ))}
        </div>
      )}

      {/* Delete Confirmation */}
      {deleteConfirmId && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-coal-950/80 p-4 backdrop-blur-sm">
          <div className="plate w-full max-w-md p-6">
            <h3 className="mb-1 font-display text-lg font-semibold text-ivory">
              Delete this record?
            </h3>
            <p className="mb-6 text-sm text-sand-400">
              This action cannot be undone. The record will be permanently
              destroyed.
            </p>
            <div className="flex justify-end gap-3">
              <button
                onClick={() => setDeleteConfirmId(null)}
                className="lk-btn lk-btn--ghost"
              >
                Cancel
              </button>
              <button onClick={confirmDelete} className="lk-btn lk-btn--danger">
                Destroy Record
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Modals */}
      <UnlockModal
        isOpen={isUnlockModalOpen}
        onUnlock={handleUnlockSuccess}
        onCancel={() => {
          setIsUnlockModalOpen(false);
          setPendingAction(null);
        }}
      />

      <ItemTypeSelector
        isOpen={isTypeSelectorOpen}
        onSelect={handleTypeSelect}
        onClose={() => setIsTypeSelectorOpen(false)}
      />

      <PasswordItemModal
        isOpen={isPasswordModalOpen}
        item={editingItem}
        decryptedSecret={editingSecret}
        onClose={() => {
          setIsPasswordModalOpen(false);
          setEditingItem(null);
          setEditingSecret(null);
        }}
        onSave={handleSavePassword}
      />
    </div>
  );
}