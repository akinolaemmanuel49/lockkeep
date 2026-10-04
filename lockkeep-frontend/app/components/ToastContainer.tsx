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
  const styles = {
    error: {
      lamp: "bg-vermillion-400",
      accent: "border-vermillion-500/40 text-vermillion-300",
      icon: "text-vermillion-300",
    },
    success: {
      lamp: "bg-patina-400",
      accent: "border-patina-500/40 text-patina-300",
      icon: "text-patina-300",
    },
    info: {
      lamp: "bg-brass-400",
      accent: "border-brass-500/40 text-brass-300",
      icon: "text-brass-300",
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

  const style = styles[toast.type];

  return (
    <div
      className={`flex max-w-sm items-start gap-3 rounded-lg border bg-coal-900 px-4 py-3 shadow-lg shadow-black/40 ${style.accent}`}
      role="alert"
    >
      <span
        className={`mt-1.5 inline-block h-1.5 w-1.5 shrink-0 rounded-full ${style.lamp}`}
      />
      <p className="flex-1 text-sm leading-relaxed">{toast.message}</p>
      <button
        onClick={onDismiss}
        className={`shrink-0 ${style.icon} cursor-pointer opacity-60 transition-opacity hover:opacity-100`}
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