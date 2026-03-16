<script lang="ts">
  import { fade, fly } from "svelte/transition";
  import { flip } from "svelte/animate";
  import { toaster, type ToastKind } from "../stores/toast.svelte";

  function getTypeStyle(kind: ToastKind) {
    switch (kind) {
      case "info":
        return "bg-primary/20 border-primary/50 text-primary";
      case "success":
        return "bg-green-500/20 border-green-500/50 text-green-400";
      case "warning":
        return "bg-yellow-500/20 border-yellow-500/50 text-yellow-400";
      case "error":
        return "bg-red-500/20 border-red-500/50 text-red-400";
    }
  }
</script>

{#snippet typeIcon(kind: ToastKind)}
  {#if kind === "info"}
    <i class="i-ri-info-i"></i>
  {:else if kind === "success"}
    <i class="i-ri-check-line"></i>
  {:else if kind === "warning"}
    <i class="i-ri-error-warning-line"></i>
  {:else if kind === "error"}
    <i class="i-ri-error-warning-fill"></i>
  {/if}
{/snippet}

<div class="pointer-events-none fixed top-4 right-4 z-50 flex flex-col gap-2">
  {#each toaster.toasts as toast (toast.id)}
    <div
      in:fly={{ x: 100, duration: 300 }}
      out:fade={{ duration: 200 }}
      animate:flip={{ duration: 200 }}
      class="pointer-events-auto flex max-w-[400px] min-w-[300px] items-center gap-3 rounded-lg border px-4 py-3 backdrop-blur-sm {getTypeStyle(
        toast.kind,
      )}">
      <span class="flex items-center font-sans text-lg">
        {@render typeIcon(toast.kind)}
      </span>
      <p class="flex-1 font-sans text-sm">{toast.message}</p>
      <button
        onclick={() => toaster.remove(toast.id)}
        class="flex items-center font-mono text-lg opacity-50 transition-opacity hover:opacity-100">
        <span class="sr-only">Dismiss {toast.kind} toast ({toast.id})</span>
        <i class="i-ri-close-line"></i>
      </button>
    </div>
  {/each}
</div>
