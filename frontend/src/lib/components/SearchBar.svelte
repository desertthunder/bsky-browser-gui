<script lang="ts">
  interface Props {
    query: string;
    source: string;
    pageSize: number;
    onSearch: (query: string, source: string) => void;
  }

  let { query = $bindable(), source = $bindable(), pageSize = $bindable(), onSearch }: Props = $props();

  const sources = [
    { value: "", label: "All" },
    { value: "saved", label: "Saved" },
    { value: "liked", label: "Liked" },
  ];
  const pageSizes = [25, 50, 100];

  function handleSubmit(e: Event) {
    e.preventDefault();
    onSearch(query, source);
  }

  function handleClick(src: string) {
    source = src;
    onSearch(query, source);
  }

  function handlePageSizeChange() {
    onSearch(query, source);
  }
</script>

<form onsubmit={handleSubmit} class="flex flex-wrap items-center gap-3">
  <div class="relative flex-1">
    <input
      id="search-posts"
      type="search"
      placeholder="Search posts..."
      bind:value={query}
      enterkeyhint="search"
      class="border-outline text-muted w-full rounded-lg border bg-black px-4 py-2.5 font-mono text-sm placeholder-[#333] focus:border-[#333] focus:outline-none" />
    <div class="absolute top-1/2 right-3 -translate-y-1/2">
      <kbd class="text-muted bg-surface border-outline rounded border px-2 py-0.5 font-mono text-xs">⌘K</kbd>
    </div>
  </div>

  <div class="bg-surface border-outline flex rounded-lg border p-0.5">
    {#each sources as s}
      <button
        type="button"
        onclick={() => {
          handleClick(s.value);
        }}
        class="rounded px-3 py-1.5 font-sans text-xs transition-colors {source === s.value
          ? 'bg-outline text-bright'
          : 'text-muted hover:text-bright'}
        ">
        {s.label}
      </button>
    {/each}
  </div>

  <button
    type="submit"
    class="bg-surface border-outline hover:bg-outline text-bright rounded-lg border px-4 py-2 font-sans text-sm transition-colors">
    Search
  </button>

  <label class="text-muted flex items-center gap-2 font-sans text-xs tracking-[0.12em] uppercase">
    <span>Results</span>
    <select
      bind:value={pageSize}
      onchange={handlePageSizeChange}
      class="border-outline bg-surface text-bright rounded-lg border px-3 py-2 font-mono text-sm focus:border-[#333] focus:outline-none">
      {#each pageSizes as size}
        <option value={size}>{size}</option>
      {/each}
    </select>
  </label>
</form>
