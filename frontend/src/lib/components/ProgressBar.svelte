<script lang="ts">
  import { slide } from "svelte/transition";
  import type { IndexStats } from "../types";

  const { isIndexing, indexStats } = $props<{ isIndexing: boolean; indexStats: IndexStats }>();
</script>

<div class="border-t border-outline bg-surface px-6 py-3" transition:slide={{ duration: 300 }}>
  <div class="flex items-center justify-between mb-2">
    <span class="font-sans text-sm text-muted">
      {#if isIndexing}
        <span class="animate-pulse">Indexing...</span>
      {:else}
        <span class="flex items-center gap-2">
          <i class="i-ri-check-line text-emerald-400"></i>
          <span>Indexing complete</span>
        </span>
      {/if}
    </span>
    <span class="font-mono text-xs text-muted">
      {indexStats.inserted} inserted / {indexStats.fetched} fetched
      {#if indexStats.errors > 0}
        <span class="text-red-500">({indexStats.errors} errors)</span>
      {/if}
    </span>
  </div>

  <div class="w-full h-1 bg-black rounded-full overflow-hidden">
    <div
      class="h-full bg-[#333] transition-all duration-300 ease-out"
      style="width: {indexStats.fetched > 0 ? (indexStats.inserted / indexStats.fetched) * 100 : 0}%">
    </div>
  </div>
</div>
