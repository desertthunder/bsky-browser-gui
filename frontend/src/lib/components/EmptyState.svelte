<script lang="ts">
  import { fade } from "svelte/transition";

  let { onRefresh }: { onRefresh: () => void } = $props();
</script>

<div transition:fade={{ duration: 300 }} class="flex h-full flex-col items-center justify-center p-8 text-center">
  <div class="mb-6">
    <span class="text-muted mx-auto mb-4 flex items-center">
      <i class="i-ri-inbox-archive-fill text-2xl"></i>
    </span>
  </div>

  <h2 class="text-bright mb-2 font-serif text-2xl">No posts indexed</h2>

  <p class="text-muted mb-6 max-w-md font-sans">
    Your Bluesky bookmarks and likes haven't been indexed yet. Click the button below to start indexing your posts.
  </p>

  <button
    onclick={onRefresh}
    class="bg-surface border-outline hover:bg-outline text-bright flex items-center gap-2 rounded-lg border px-6 py-3 font-sans transition-colors">
    <i class="i-ri-refresh-line h-5 w-5"></i>
    <span>Refresh Data</span>
  </button>
</div>
