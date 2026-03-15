export type ToastKind = "info" | "success" | "warning" | "error";

export type Toast = { id: number; message: string; kind: ToastKind };

class ToastStore {
  toasts = $state<Toast[]>([]);
  private id = 0;

  add(message: string, kind: ToastKind, duration = 5000) {
    const toastId = ++this.id;
    this.toasts = [...this.toasts, { id: toastId, message, kind }];

    setTimeout(() => {
      this.remove(toastId);
    }, duration);

    return toastId;
  }

  remove(id: number) {
    this.toasts = this.toasts.filter((t) => t.id !== id);
  }

  info(message: string, duration?: number) {
    return this.add(message, "info", duration);
  }

  success(message: string, duration?: number) {
    return this.add(message, "success", duration);
  }

  warning(message: string, duration?: number) {
    return this.add(message, "warning", duration);
  }

  error(message: string, duration?: number) {
    return this.add(message, "error", duration || 8000);
  }
}

export const toaster = new ToastStore();
