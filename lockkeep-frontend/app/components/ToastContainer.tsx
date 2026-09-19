import { useToast, type ToastPosition } from "~/providers/toast";

const positions: ToastPosition[] = [
  "top",
  "top-left",
  "top-right",
  "bottom",
  "bottom-left",
  "bottom-right",
];

const positionClasses: Record<ToastPosition, string> = {
  top: "top-6 left-1/2 -translate-x-1/2",
  "top-left": "top-6 left-6",
  "top-right": "top-6 right-6",
  bottom: "bottom-6 left-1/2 -translate-x-1/2",
  "bottom-left": "bottom-6 left-6",
  "bottom-right": "bottom-6 right-6",
};

export default function ToastContainer() {
  const { toasts, removeToast } = useToast();

  if (toasts.length === 0) return null;

  return (
    <>
      {positions.map((position) => {
        const items = toasts.filter((t) => t.position === position);

        if (items.length === 0) return null;

        return (
          <div
            key={position}
            className={`fixed z-[100] flex flex-col gap-3 ${positionClasses[position]}`}
          >
            {items.map((toast) => (
              <ToastItem
                key={toast.id}
                toast={toast}
                onDismiss={() => removeToast(toast.id)}
              />
            ))}
          </div>
        );
      })}
    </>
  );
}

function ToastItem({
  toast,
  onDismiss,
}: {
  toast: { id: string; message: string; type: "error" | "success" | "info" };
  onDismiss: () => void;
}) {
  const colors = {
    error: {
      wrapper:
        "border-red-400/30 bg-red-500 text-red-50 shadow-lg shadow-red-500/20",
      icon: "text-red-100",
    },
    success: {
      wrapper:
        "border-green-400/30 bg-green-500 text-green-50 shadow-lg shadow-green-500/20",
      icon: "text-green-100",
    },
    info: {
      wrapper:
        "border-sky-400/30 bg-sky-500 text-sky-50 shadow-lg shadow-sky-500/20",
      icon: "text-sky-100",
    },
  };

  const icons = {
    error: (
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
        <circle cx="12" cy="12" r="10" />
        <line x1="12" y1="8" x2="12" y2="12" />
        <line x1="12" y1="16" x2="12.01" y2="16" />
      </svg>
    ),
    success: (
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
        <polyline points="20 6 9 17 4 12" />
      </svg>
    ),
    info: (
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
        <circle cx="12" cy="12" r="10" />
        <line x1="12" y1="16" x2="12" y2="12" />
        <line x1="12" y1="8" x2="12.01" y2="8" />
      </svg>
    ),
  };

  const color = colors[toast.type];

  return (
    <div
      className={`flex max-w-sm items-start gap-3 rounded-lg border px-4 py-3 shadow-lg ${color.wrapper} animate-in slide-in-from-right fade-in duration-300`}
      role="alert"
    >
      <div className={`mt-0.5 shrink-0 ${color.icon}`}>{icons[toast.type]}</div>
      <p className="flex-1 text-sm leading-relaxed">{toast.message}</p>
      <button
        onClick={onDismiss}
        className="shrink-0 text-current opacity-60 hover:opacity-100 cursor-pointer"
        aria-label="Dismiss"
      >
        <svg
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <line x1="18" y1="6" x2="6" y2="18" />
          <line x1="6" y1="6" x2="18" y2="18" />
        </svg>
      </button>
    </div>
  );
}
