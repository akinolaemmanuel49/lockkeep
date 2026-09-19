import {
  createContext,
  useContext,
  useState,
  useCallback,
  type ReactNode,
} from "react";

export type ToastPosition =
  | "top"
  | "top-left"
  | "top-right"
  | "bottom"
  | "bottom-left"
  | "bottom-right";

export interface Toast {
  id: string;
  message: string;
  type: "error" | "success" | "info";
  position: ToastPosition;
}

interface ToastContextValue {
  toasts: Toast[];
  addToast: (
    message: string,
    type?: Toast["type"],
    position?: Toast["position"],
  ) => void;
  removeToast: (id: string) => void;
}

const ToastContext = createContext<ToastContextValue | null>(null);

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const addToast = useCallback(
    (
      message: string,
      type: Toast["type"] = "info",
      position: Toast["position"] = "top-right",
    ) => {
      const id = crypto.randomUUID();
      setToasts((prev) => [...prev, { id, message, type, position }]);

      // Auto-dismiss after 5 seconds
      setTimeout(() => {
        setToasts((prev) => prev.filter((t) => t.id !== id));
      }, 5000);
    },
    [],
  );

  const removeToast = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  return (
    <ToastContext.Provider value={{ toasts, addToast, removeToast }}>
      {children}
    </ToastContext.Provider>
  );
}

export function useToast(): ToastContextValue {
  const ctx = useContext(ToastContext);
  if (!ctx) throw new Error("useToast must be used within ToastProvider");
  return ctx;
}
