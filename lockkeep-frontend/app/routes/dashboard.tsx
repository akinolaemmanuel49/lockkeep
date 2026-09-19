import { useState } from "react";
import { useVault } from "~/providers/vault";
import { useToast } from "~/providers/toast";
import VaultItemCard from "~/components/VaultItemCard";
import PasswordItemModal from "~/components/PasswordItemModal";
import ItemTypeSelector from "~/components/ItemTypeSelector";
import UnlockModal from "~/components/UnlockModal";
import type { VaultItem, VaultItemType } from "~/types/index";
import { requireAuth } from "~/lib/auth-guard";
import type { Route } from "./+types/dashboard";

export const clientLoader = () => {
  return requireAuth();
};

export function meta({ }: Route.MetaArgs) {
  return [
    { title: "Dashboard - LockKeep" },
    {
      name: "description",
      content: "Manage your encrypted vault items. Search, add, and organize your secrets securely.",
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
        addToast(err instanceof Error ? err.message : "Failed to decrypt secret", "error");
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
      {/* Toolbar */}
      <div className="mb-6 flex flex-wrap items-center gap-4">
        <div className="relative flex-1 min-w-[280px]">
          <svg
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="#475569"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            className="absolute left-3 top-1/2 -translate-y-1/2"
          >
            <circle cx="11" cy="11" r="8" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search by name, site, or identifier..."
            className="w-full rounded-lg border border-slate-800 bg-slate-900 py-2.5 pl-10 pr-4 text-sm text-slate-200 placeholder:text-slate-600 focus:border-sky-400 focus:outline-none"
          />
          {searchQuery && (
            <button
              onClick={() => setSearchQuery("")}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-600 hover:text-slate-300"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <line x1="18" y1="6" x2="6" y2="18" />
                <line x1="6" y1="6" x2="18" y2="18" />
              </svg>
            </button>
          )}
        </div>

        <button
          onClick={handleAddClick}
          className="flex items-center gap-2 rounded-lg bg-sky-400 px-4 py-2.5 text-sm font-semibold text-slate-950 hover:bg-sky-300 cursor-pointer"
        >
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <line x1="12" y1="5" x2="12" y2="19" />
            <line x1="5" y1="12" x2="19" y2="12" />
          </svg>
          Add Item
        </button>
      </div>

      {/* Content */}
      {isLoading && items.length === 0 ? (
        <div className="flex items-center justify-center gap-3 py-20 text-slate-500">
          <div className="h-5 w-5 animate-spin rounded-full border-2 border-slate-700 border-t-sky-400" />
          <span>Loading vault items...</span>
        </div>
      ) : filtered.length === 0 ? (
        <div className="flex flex-col items-center py-20 text-center">
          <svg
            width="48"
            height="48"
            viewBox="0 0 24 24"
            fill="none"
            stroke="#334155"
            strokeWidth="1.5"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
            <path d="M7 11V7a5 5 0 0 1 10 0v4" />
          </svg>
          <p className="mt-4 text-lg font-semibold text-slate-600">
            {searchQuery ? "No matching items" : "No items yet"}
          </p>
          <p className="mt-1 text-sm text-slate-700">
            {searchQuery
              ? "Try adjusting your search"
              : "Click Add Item to store your first password"}
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
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
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4">
          <div className="w-full max-w-md rounded-2xl border border-slate-700 bg-slate-900 p-6">
            <h3 className="mb-2 text-lg font-semibold text-slate-100">Delete Item?</h3>
            <p className="mb-6 text-sm text-slate-400">
              This action cannot be undone. The item will be permanently removed.
            </p>
            <div className="flex justify-end gap-3">
              <button
                onClick={() => setDeleteConfirmId(null)}
                className="rounded-lg border border-slate-700 px-4 py-2 text-sm font-medium text-slate-400 hover:text-slate-200"
              >
                Cancel
              </button>
              <button
                onClick={confirmDelete}
                className="rounded-lg bg-red-500 px-4 py-2 text-sm font-semibold text-white hover:bg-red-400"
              >
                Delete
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