<script lang="ts">
  import { fade, fly } from "svelte/transition";
  import { flip } from "svelte/animate";
  import { toaster, type ToastKind } from "../stores/toast.svelte";

  function getTypeStyle(kind: ToastKind) {
    switch (kind) {
      case "info":
        return "bg-blue-500/20 border-blue-500/50 text-blue-400";
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

<div class="fixed top-4 right-4 z-50 flex flex-col gap-2 pointer-events-none">
  {#each toaster.toasts as toast (toast.id)}
    <div
      in:fly={{ x: 100, duration: 300 }}
      out:fade={{ duration: 200 }}
      animate:flip={{ duration: 200 }}
      class="pointer-events-auto flex items-center gap-3 px-4 py-3 rounded-lg border backdrop-blur-sm min-w-[300px] max-w-[400px] {getTypeStyle(
        toast.kind,
      )}">
      <span class="font-sans text-lg flex items-center">
        {@render typeIcon(toast.kind)}
      </span>
      <p class="font-sans text-sm flex-1">{toast.message}</p>
      <button
        onclick={() => toaster.remove(toast.id)}
        class="font-mono text-lg opacity-50 hover:opacity-100 transition-opacity flex items-center">
        <span class="sr-only">Dismiss {toast.kind} toast ({toast.id})</span>
        <i class="i-ri-close-line"></i>
      </button>
    </div>
  {/each}
</div>
