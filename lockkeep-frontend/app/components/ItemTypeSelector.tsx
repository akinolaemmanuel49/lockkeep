import type { VaultItemType } from "../types/index";

interface ItemTypeSelectorProps {
    isOpen: boolean;
    onSelect: (type: VaultItemType) => void;
    onClose: () => void;
}

const ITEM_TYPES: Array<{
    type: VaultItemType;
    label: string;
    description: string;
    icon: React.ReactNode;
    available: boolean;
}> = [
        {
            type: "login",
            label: "Password / Login",
            description: "Store usernames, passwords, and site URLs",
            available: true,
            icon: (
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                    <rect x="3" y="11" width="18" height="11" rx="2" />
                    <path d="M7 11V7a5 5 0 0 1 10 0v4" />
                </svg>
            ),
        },
        {
            type: "environment",
            label: "Environment Variable",
            description: "Store API keys, tokens, and env secrets",
            available: false,
            icon: (
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                    <polyline points="4 17 10 11 4 5" />
                    <line x1="12" y1="19" x2="20" y2="19" />
                </svg>
            ),
        },
        {
            type: "ssh_key",
            label: "SSH Key",
            description: "Store private SSH keys securely",
            available: false,
            icon: (
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                    <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4" />
                </svg>
            ),
        },
        {
            type: "secure_note",
            label: "Secure Note",
            description: "Encrypted text notes",
            available: false,
            icon: (
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                    <polyline points="14 2 14 8 20 8" />
                </svg>
            ),
        },
        {
            type: "payment_card",
            label: "Payment Card",
            description: "Store card numbers and billing info",
            available: false,
            icon: (
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                    <rect x="1" y="4" width="22" height="16" rx="2" />
                    <line x1="1" y1="10" x2="23" y2="10" />
                </svg>
            ),
        },
        {
            type: "api_key",
            label: "API Key",
            description: "Store service API keys",
            available: false,
            icon: (
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                    <circle cx="12" cy="12" r="3" />
                    <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" />
                </svg>
            ),
        },
    ];

export default function ItemTypeSelector({ isOpen, onSelect, onClose }: ItemTypeSelectorProps) {
    if (!isOpen) return null;

    return (
        <div
            className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4"
            onClick={onClose}
        >
            <div
                className="w-full max-w-lg rounded-2xl border border-slate-700 bg-slate-900 p-6"
                onClick={(e) => e.stopPropagation()}
            >
                <div className="mb-4 flex items-center justify-between">
                    <h2 className="text-lg font-semibold text-slate-100">Add Vault Item</h2>
                    <button
                        onClick={onClose}
                        className="rounded-lg p-1 text-slate-500 hover:text-slate-300"
                    >
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                            <line x1="18" y1="6" x2="6" y2="18" />
                            <line x1="6" y1="6" x2="18" y2="18" />
                        </svg>
                    </button>
                </div>

                <p className="mb-4 text-sm text-slate-500">
                    Select the type of item you want to store in your vault.
                </p>

                <div className="grid grid-cols-1 gap-3">
                    {ITEM_TYPES.map((item) => (
                        <button
                            key={item.type}
                            onClick={() => item.available && onSelect(item.type)}
                            disabled={!item.available}
                            className={`flex items-center gap-4 rounded-xl border p-4 text-left transition-colors ${item.available
                                ? "border-slate-700 bg-slate-950 hover:border-sky-400 hover:bg-slate-900"
                                : "border-slate-800 bg-slate-950/50 opacity-50 cursor-not-allowed"
                                }`}
                        >
                            <div className={`shrink-0 ${item.available ? "text-sky-400" : "text-slate-600"}`}>
                                {item.icon}
                            </div>
                            <div className="flex-1 min-w-0">
                                <div className="flex items-center gap-2">
                                    <span className="font-medium text-slate-200">{item.label}</span>
                                    {!item.available && (
                                        <span className="rounded-full bg-slate-800 px-2 py-0.5 text-[10px] font-medium uppercase tracking-wider text-slate-500">
                                            Soon
                                        </span>
                                    )}
                                </div>
                                <p className="mt-0.5 text-sm text-slate-500">{item.description}</p>
                            </div>
                            {item.available && (
                                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="shrink-0 text-slate-500">
                                    <polyline points="9 18 15 12 9 6" />
                                </svg>
                            )}
                        </button>
                    ))}
                </div>
            </div>
        </div>
    );
}