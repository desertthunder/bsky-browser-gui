<script lang="ts">
  import { slide } from "svelte/transition";
  import type { IndexStats } from "../types";

  type Props = { isIndexing: boolean; indexStats: IndexStats };

  const { isIndexing, indexStats }: Props = $props();
</script>

<div class="border-outline bg-surface border-t px-6 py-3" transition:slide={{ duration: 300 }}>
  <div class="mb-2 flex items-center justify-between">
    <span class="text-muted font-sans text-sm">
      {#if isIndexing}
        <span class="animate-pulse">Indexing...</span>
      {:else}
        <span class="flex items-center gap-2">
          <i class="i-ri-check-line text-emerald-400"></i>
          <span>Indexing complete</span>
        </span>
      {/if}
    </span>
    <span class="text-muted font-mono text-xs">
      {indexStats.inserted} inserted / {indexStats.fetched} fetched
      {#if indexStats.errors > 0}
        <span class="text-red-500">({indexStats.errors} errors)</span>
      {/if}
    </span>
  </div>

  <div class="h-1 w-full overflow-hidden rounded-full bg-black">
    {#if isIndexing}
      <div class="h-full w-1/3 animate-[progress-indeterminate_1.2s_ease-in-out_infinite] bg-[#555]"></div>
    {:else}
      <div class="h-full bg-[#333] transition-all duration-300 ease-out" style="width: 100%"></div>
    {/if}
  </div>
</div>
