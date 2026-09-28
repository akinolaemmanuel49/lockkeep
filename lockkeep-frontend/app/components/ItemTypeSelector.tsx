import type { VaultItemType } from "../types/index";
import { RECORD_TYPES } from "../lib/recordTypes";
import Nameplate from "./Nameplate";

interface ItemTypeSelectorProps {
  isOpen: boolean;
  onSelect: (type: VaultItemType) => void;
  onClose: () => void;
}

export default function ItemTypeSelector({
  isOpen,
  onSelect,
  onClose,
}: ItemTypeSelectorProps) {
  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-coal-950/85 p-4 backdrop-blur-sm"
      onClick={onClose}
    >
      <div className="plate w-full max-w-lg p-6" onClick={(e) => e.stopPropagation()}>
        <div className="mb-4 flex items-center justify-between border-b border-brass-500/15 pb-4">
          <div>
            <Nameplate>Classification</Nameplate>
            <h2 className="mt-1.5 font-display text-xl font-medium text-ivory">
              Choose a Record Type
            </h2>
          </div>
          <button
            onClick={onClose}
            className="rounded-lg p-1.5 text-sand-600 transition-colors hover:text-ivory"
            aria-label="Close"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>

        <p className="mb-4 text-sm text-sand-500">
          Select the type of item you want to store in your vault.
        </p>

        <div className="grid grid-cols-1 gap-3">
          {RECORD_TYPES.map((item) => (
            <button
              key={item.type}
              onClick={() => onSelect(item.type)}
              className="group flex items-center gap-4 rounded-lg border border-brass-500/20 bg-coal-950 p-4 text-left transition-colors hover:border-brass-400/50 hover:bg-coal-850"
            >
              <div className="shrink-0 text-brass-400">{item.icon}</div>
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                  <span className="font-medium text-ivory">{item.label}</span>
                </div>
                <p className="mt-0.5 text-sm text-sand-500">{item.description}</p>
              </div>
              <svg
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                className="shrink-0 text-brass-500 transition-transform group-hover:translate-x-0.5"
              >
                <polyline points="9 18 15 12 9 6" />
              </svg>
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}