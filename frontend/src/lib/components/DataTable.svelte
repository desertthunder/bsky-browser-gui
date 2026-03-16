<script lang="ts">
  import type { main } from "../../../wailsjs/go/models";
  import { formatShortDate } from "../date";
  import PostText from "./PostText.svelte";

  interface Props {
    posts: main.SearchResult[];
    totalPosts: number;
    currentPage: number;
    pageSize: number;
    sortColumn: string;
    sortDirection: "asc" | "desc";
    onSort: (column: string) => void;
    onPageChange: (page: number) => void;
    onOpenPost: (post: main.SearchResult) => void;
    selectedPostURI?: string | null;
  }

  let {
    posts,
    totalPosts,
    currentPage,
    pageSize,
    sortColumn,
    sortDirection,
    onSort,
    onPageChange,
    onOpenPost,
    selectedPostURI = null,
  }: Props = $props();

  const columns = [
    { key: "author_handle", label: "Author", width: "w-36" },
    { key: "text", label: "Text", width: "min-w-[32rem]" },
    { key: "created_at", label: "Created", width: "w-36" },
    { key: "like_count", label: "LIKE", width: "w-20" },
    { key: "repost_count", label: "REPOST", width: "w-20" },
    { key: "reply_count", label: "REPLY", width: "w-20" },
    { key: "source", label: "Source", width: "w-28" },
  ];

  let totalPages = $derived(Math.max(1, Math.ceil(totalPosts / pageSize)));
  let pageStart = $derived(totalPosts === 0 ? 0 : (currentPage - 1) * pageSize + 1);
  let pageEnd = $derived(Math.min(pageStart + posts.length - 1, totalPosts));
  let visiblePages = $derived.by(() => {
    const pages: number[] = [];
    const start = Math.max(1, currentPage - 2);
    const end = Math.min(totalPages, currentPage + 2);

    for (let page = start; page <= end; page += 1) {
      pages.push(page);
    }

    return pages;
  });
</script>

{#snippet columnLabel(label: string)}
  {#if label === "LIKE"}
    <span class="flex-items-center">
      <i class="i-ri-heart-line text-red-500"></i>
    </span>
  {:else if label === "REPOST"}
    <span class="flex-items-center">
      <i class="i-ri-repeat-line text-blue-500"></i>
    </span>
  {:else if label === "REPLY"}
    <span class="flex-items-center">
      <i class="i-ri-message-2-line text-green-500"></i>
    </span>
  {:else}
    <span>{label}</span>
  {/if}
{/snippet}

{#snippet sortIcon(column: string)}
  <span class="flex items-center">
    {#if sortColumn !== column}
      <i class="i-ri-arrow-up-down-line"></i>
    {:else if sortDirection === "asc"}
      <i class="i-ri-arrow-up-line"></i>
    {:else}
      <i class="i-ri-arrow-down-line"></i>
    {/if}
  </span>
{/snippet}

<div
  class="border-outline bg-surface flex h-full min-h-0 flex-col overflow-hidden rounded-[1.25rem] border shadow-[0_18px_60px_rgba(0,0,0,0.35)]">
  <div class="min-h-0 flex-1 overflow-auto">
    <table class="w-full min-w-296 border-separate border-spacing-0">
      <thead class="sticky top-0 z-10 bg-black/95 backdrop-blur">
        <tr>
          {#each columns as column}
            <th
              class="border-outline text-muted hover:text-bright cursor-pointer border-b px-4 py-3 text-left font-sans text-xs tracking-[0.16em] uppercase select-none {column.width}"
              onclick={() => onSort(column.key)}>
              <div class="flex items-center gap-1">
                {@render columnLabel(column.label)}
                {@render sortIcon(column.key)}
              </div>
            </th>
          {/each}
        </tr>
      </thead>

      <tbody class="divide-outline divide-y">
        {#each posts as post}
          <tr
            class="group cursor-pointer transition-colors {selectedPostURI === post.uri
              ? 'bg-outline/80'
              : 'hover:bg-outline/50'}"
            onclick={() => onOpenPost(post)}>
            <td class="text-muted truncate px-4 py-3 font-mono text-xs">
              @{post.author_handle}
            </td>

            <td class="text-bright px-4 py-3 font-mono text-sm">
              <div class="line-clamp-2">
                <PostText text={post.text} facetsJson={post.facets} maxLength={120} />
              </div>
            </td>

            <td class="text-muted px-4 py-3 font-mono text-xs">
              {formatShortDate(post.created_at)}
            </td>

            <td class="text-bright px-4 py-3 text-center font-mono text-xs">
              {post.like_count || 0}
            </td>

            <td class="text-bright px-4 py-3 text-center font-mono text-xs">
              {post.repost_count || 0}
            </td>

            <td class="text-bright px-4 py-3 text-center font-mono text-xs">
              {post.reply_count || 0}
            </td>

            <td class="px-4 py-3">
              <span
                class="rounded-full px-2 py-0.5 font-sans text-xs {post.source === 'saved'
                  ? 'bg-primary/20 text-primary'
                  : 'bg-secondary/20 text-secondary'}">
                {post.source}
              </span>
            </td>
          </tr>
        {:else}
          <tr>
            <td colspan={columns.length} class="px-4 py-12 text-center">
              <p class="font-sans text-muted">No posts found</p>
              <p class="mt-2 font-mono text-xs text-[#333]">Try searching or refreshing your data</p>
            </td>
          </tr>
        {/each}
      </tbody>

      {#if posts.length > 0}
        <tfoot class="sticky bottom-0 z-10 bg-black/95 backdrop-blur">
          <tr>
            <td colspan={columns.length} class="border-outline border-t px-4 py-3">
              <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                <p class="text-muted font-mono text-xs tracking-[0.14em] uppercase">
                  Showing {pageStart}-{pageEnd} of {totalPosts}
                </p>

                <div class="flex flex-wrap items-center gap-2">
                  <button
                    type="button"
                    class="border-outline text-muted hover:text-bright rounded-full border px-3 py-1.5 font-mono text-xs transition-colors disabled:opacity-40"
                    onclick={() => onPageChange(Math.max(1, currentPage - 1))}
                    disabled={currentPage === 1}>
                    Prev
                  </button>

                  {#each visiblePages as page}
                    <button
                      type="button"
                      class="min-w-9 rounded-full border px-3 py-1.5 font-mono text-xs transition-colors {page ===
                      currentPage
                        ? 'border-primary bg-primary/15 text-primary'
                        : 'border-outline text-muted hover:text-bright'}"
                      onclick={() => onPageChange(page)}>
                      {page}
                    </button>
                  {/each}

                  <button
                    type="button"
                    class="border-outline text-muted hover:text-bright rounded-full border px-3 py-1.5 font-mono text-xs transition-colors disabled:opacity-40"
                    onclick={() => onPageChange(Math.min(totalPages, currentPage + 1))}
                    disabled={currentPage === totalPages}>
                    Next
                  </button>
                </div>
              </div>
            </td>
          </tr>
        </tfoot>
      {/if}
    </table>
  </div>
</div>
