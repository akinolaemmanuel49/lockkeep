import type { ReactNode } from "react";
import type { VaultItemType } from "~/types/index";

export interface RecordMetadataField {
  key: string;
  label: string;
  kind?: "text" | "url" | "tel" | "month";
  placeholder?: string;
  required?: boolean;
}

export interface RecordTypeDefinition {
  type: VaultItemType;
  label: string;
  description: string;
  icon: ReactNode;
  secret: {
    label: string;
    placeholder?: string;
    multiline?: boolean;
    generate?: boolean;
  };
  metadataFields: RecordMetadataField[];
}

export const RECORD_TYPES: RecordTypeDefinition[] = [
  {
    type: "login",
    label: "Password / Login",
    description: "Store usernames, passwords, and site URLs",
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
        <rect x="3" y="11" width="18" height="11" rx="2" />
        <path d="M7 11V7a5 5 0 0 1 10 0v4" />
      </svg>
    ),
    secret: {
      label: "Password",
      placeholder: "Enter or generate password",
      generate: true,
    },
    metadataFields: [
      { key: "siteUrl", label: "Site URL", kind: "url", placeholder: "https://..." },
      { key: "identifier", label: "Identifier (Email / Username)", required: true, placeholder: "user@example.com" },
    ],
  },
  {
    type: "environment",
    label: "Environment Variable",
    description: "Store environment keys, tokens, and configuration values",
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
        <polyline points="4 17 10 11 4 5" />
        <line x1="12" y1="19" x2="20" y2="19" />
      </svg>
    ),
    secret: {
      label: "Value",
      placeholder: "The value of the environment variable",
    },
    metadataFields: [],
  },
  {
    type: "ssh_key",
    label: "SSH Key",
    description: "Store private SSH keys securely",
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
        <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4" />
      </svg>
    ),
    secret: {
      label: "Private Key",
      placeholder: "-----BEGIN OPENSSH PRIVATE KEY-----\n...",
      multiline: true,
    },
    metadataFields: [
      { key: "username", label: "Username / Comment", placeholder: "root@server" },
    ],
  },
  {
    type: "secure_note",
    label: "Secure Note",
    description: "Encrypted text notes, fragments, and one-off secrets",
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
        <polyline points="14 2 14 8 20 8" />
      </svg>
    ),
    secret: {
      label: "Note",
      placeholder: "Write the encrypted note content...",
      multiline: true,
    },
    metadataFields: [],
  },
  {
    type: "payment_card",
    label: "Payment Card",
    description: "Store card numbers and billing details",
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
        <rect x="1" y="4" width="22" height="16" rx="2" />
        <line x1="1" y1="10" x2="23" y2="10" />
      </svg>
    ),
    secret: {
      label: "Card Number",
      placeholder: "•••• •••• •••• ••••",
    },
    metadataFields: [
      { key: "cardHolder", label: "Card Holder", placeholder: "Name on card" },
      { key: "expiresAt", label: "Expiration (MM/YYYY)", kind: "month" },
    ],
  },
  {
    type: "api_key",
    label: "API Key",
    description: "Store service API keys and secrets",
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
        <circle cx="12" cy="12" r="3" />
        <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" />
      </svg>
    ),
    secret: {
      label: "Key",
      placeholder: "Paste the API key value",
    },
    metadataFields: [
      { key: "issuer", label: "Issuer / Service", placeholder: "e.g., Stripe, OpenAI" },
      { key: "scopes", label: "Scopes / Permissions", placeholder: "read, write" },
    ],
  },
];

export function getRecordType(type: VaultItemType): RecordTypeDefinition {
  return (
    RECORD_TYPES.find((t) => t.type === type) ?? {
      type,
      label: type.replace("_", " ").replace(/\b\w/g, (c) => c.toUpperCase()),
      description: "",
      icon: null,
      secret: { label: "Secret" },
      metadataFields: [],
    }
  );
}