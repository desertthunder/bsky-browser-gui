<script lang="ts">
  interface Props {
    query: string;
    source: string;
    onSearch: (query: string, source: string) => void;
  }

  let { query = $bindable(), source = $bindable(), onSearch }: Props = $props();

  const sources = [
    { value: "", label: "All" },
    { value: "saved", label: "Saved" },
    { value: "liked", label: "Liked" },
  ];

  function handleSubmit(e: Event) {
    e.preventDefault();
    onSearch(query, source);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Enter") {
      onSearch(query, source);
    }
  }

  function handleClick(src: string) {
    source = src;
    onSearch(query, source);
  }
</script>

<form onsubmit={handleSubmit} class="flex items-center gap-3">
  <div class="flex-1 relative">
    <input
      type="text"
      placeholder="Search posts..."
      bind:value={query}
      onkeydown={handleKeydown}
      class="w-full bg-black border border-outline rounded-lg px-4 py-2.5 font-mono text-sm text-muted placeholder-[#333] focus:outline-none focus:border-[#333]" />
    <div class="absolute right-3 top-1/2 -translate-y-1/2">
      <kbd class="font-mono text-xs text-muted bg-surface px-2 py-0.5 rounded border border-outline">⌘K</kbd>
    </div>
  </div>

  <div class="flex bg-surface rounded-lg border border-outline p-0.5">
    {#each sources as s}
      <button
        type="button"
        onclick={() => {
          handleClick(s.value);
        }}
        class="px-3 py-1.5 font-sans text-xs rounded transition-colors {source === s.value
          ? 'bg-outline text-bright'
          : 'text-muted hover:text-bright'}
        ">
        {s.label}
      </button>
    {/each}
  </div>

  <button
    type="submit"
    class="bg-surface border border-outline hover:bg-outline text-bright font-sans text-sm px-4 py-2 rounded-lg transition-colors">
    Search
  </button>
</form>
