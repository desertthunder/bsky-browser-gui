export function parseDateValue(value: unknown): Date | null {
  if (!value) {
    return null;
  }

  if (value instanceof Date) {
    return Number.isNaN(value.getTime()) || value.getUTCFullYear() <= 1 ? null : value;
  }

  if (typeof value === "string" || typeof value === "number") {
    const parsed = new Date(value);
    return Number.isNaN(parsed.getTime()) || parsed.getUTCFullYear() <= 1 ? null : parsed;
  }

  return null;
}

export function formatShortDate(value: unknown): string {
  const date = parseDateValue(value);
  if (!date) {
    return "-";
  }

  return date.toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" });
}

export function formatLongDateTime(value: unknown): string {
  const date = parseDateValue(value);
  if (!date) {
    return "Unknown date";
  }

  return date.toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}
