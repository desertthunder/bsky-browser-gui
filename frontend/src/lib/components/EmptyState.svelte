<script lang="ts">
  import { fade } from "svelte/transition";

  interface Props {
    onRefresh: () => void;
  }

  let { onRefresh }: Props = $props();
</script>

<div 
  transition:fade={{ duration: 300 }}
  class="flex flex-col items-center justify-center h-full text-center p-8"
>
  <div class="mb-6">
    <!-- Empty state icon - inbox/archive symbol -->
    <svg 
      class="w-16 h-16 text-muted mx-auto mb-4" 
      fill="none" 
      stroke="currentColor" 
      viewBox="0 0 24 24"
    >
      <path 
        stroke-linecap="round" 
        stroke-linejoin="round" 
        stroke-width="1.5"
        d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
      />
    </svg>
  </div>

  <h2 class="font-serif text-2xl text-bright mb-2">No posts indexed</h2>
  
  <p class="font-sans text-muted mb-6 max-w-md">
    Your Bluesky bookmarks and likes haven't been indexed yet. 
    Click the button below to start indexing your posts.
  </p>

  <button
    onclick={onRefresh}
    class="bg-surface border border-outline hover:bg-outline text-bright font-sans py-3 px-6 rounded-lg transition-colors flex items-center gap-2"
  >
    <!-- Refresh icon -->
    <svg 
      class="w-5 h-5" 
      fill="none" 
      stroke="currentColor" 
      viewBox="0 0 24 24"
    >
      <path 
        stroke-linecap="round" 
        stroke-linejoin="round" 
        stroke-width="2"
        d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
      />
    </svg>
    Refresh Data
  </button>
</div>
