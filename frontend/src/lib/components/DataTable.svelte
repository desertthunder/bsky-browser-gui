<script lang="ts">
  import { BrowserOpenURL } from "../../../wailsjs/runtime/runtime";
  import type { main } from "../../../wailsjs/go/models";
  import PostText from "./PostText.svelte";

  interface Props {
    posts: main.SearchResult[];
    sortColumn: string;
    sortDirection: "asc" | "desc";
    onSort: (column: string) => void;
  }

  let { posts, sortColumn, sortDirection, onSort }: Props = $props();

  const columns = [
    { key: "author_handle", label: "Author", width: "w-32" },
    { key: "text", label: "Text", width: "flex-1" },
    { key: "created_at", label: "Created", width: "w-32" },
    { key: "like_count", label: "♥", width: "w-16" },
    { key: "repost_count", label: "🔁", width: "w-16" },
    { key: "reply_count", label: "💬", width: "w-16" },
    { key: "source", label: "Source", width: "w-20" },
  ];

  function formatDate(dateStr: string): string {
    if (!dateStr) return "-";
    const date = new Date(dateStr);
    return date.toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    });
  }

  function truncateText(text: string, maxLength: number = 120): string {
    if (!text) return "";
    if (text.length <= maxLength) return text;
    return text.slice(0, maxLength) + "...";
  }

  /**
   * Convert at:// URI to bsky.app URL
   * at://did:plc:xxx/app.bsky.feed.post/xxx -> https://bsky.app/profile/did:plc:xxx/post/xxx
   */
  function buildPostURL(uri: string): string {
    const match = uri.match(/at:\/\/([^/]+)\/app\.bsky\.feed\.post\/(.+)/);
    if (match) {
      return `https://bsky.app/profile/${match[1]}/post/${match[2]}`;
    }
    return uri;
  }

  function handleRowClick(uri: string) {
    const url = buildPostURL(uri);
    BrowserOpenURL(url);
  }

  function getSortIcon(column: string): string {
    if (sortColumn !== column) return "↕";
    return sortDirection === "asc" ? "↑" : "↓";
  }
</script>

<div class="border border-outline rounded-lg overflow-hidden bg-surface">
  <div class="overflow-x-auto">
    <table class="w-full">
      <thead class="bg-black border-b border-outline">
        <tr>
          {#each columns as column}
            <th
              class="px-4 py-3 text-left font-sans text-xs text-muted cursor-pointer hover:text-bright select-none {column.width}"
              onclick={() => onSort(column.key)}>
              <div class="flex items-center gap-1">
                <span>{column.label}</span>
                <span class="font-mono text-[10px]">{getSortIcon(column.key)}</span>
              </div>
            </th>
          {/each}
        </tr>
      </thead>

      <tbody class="divide-y divide-outline">
        {#each posts as post}
          <tr class="hover:bg-black/50 cursor-pointer transition-colors group" onclick={() => handleRowClick(post.uri)}>
            <td class="px-4 py-3 font-mono text-xs text-muted truncate">
              @{post.author_handle}
            </td>

            <td class="px-4 py-3 font-mono text-sm text-bright">
              <div class="line-clamp-2">
                <PostText text={post.text} facetsJson={post.facets} maxLength={120} />
              </div>
            </td>

            <td class="px-4 py-3 font-mono text-xs text-muted">
              {formatDate(post.created_at)}
            </td>

            <td class="px-4 py-3 font-mono text-xs text-bright text-center">
              {post.like_count || 0}
            </td>

            <td class="px-4 py-3 font-mono text-xs text-bright text-center">
              {post.repost_count || 0}
            </td>

            <td class="px-4 py-3 font-mono text-xs text-bright text-center">
              {post.reply_count || 0}
            </td>

            <td class="px-4 py-3">
              <span
                class="font-sans text-xs px-2 py-0.5 rounded-full {post.source === 'saved'
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
              <p class="font-mono text-xs text-[#333] mt-2">Try searching or refreshing your data</p>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
