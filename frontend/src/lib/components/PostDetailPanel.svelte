<script lang="ts">
  import { BrowserOpenURL } from "../../../wailsjs/runtime/runtime";
  import type { main } from "../../../wailsjs/go/models";
  import PostText from "./PostText.svelte";
  import { formatLongDateTime } from "../date";

  interface Props {
    post: main.SearchResult;
    onClose: () => void;
  }

  let { post, onClose }: Props = $props();

  function buildPostURL(uri: string): string {
    const match = uri.match(/at:\/\/([^/]+)\/app\.bsky\.feed\.post\/(.+)/);
    if (match) {
      return `https://bsky.app/profile/${match[1]}/post/${match[2]}`;
    }
    return uri;
  }

  function openInBrowser() {
    BrowserOpenURL(buildPostURL(post.uri));
  }
</script>

<aside
  class="border-outline bg-surface/95 flex h-full min-h-0 w-full flex-col overflow-hidden rounded-[1.25rem] border shadow-[0_24px_80px_rgba(0,0,0,0.45)] backdrop-blur xl:w-100 xl:min-w-100">
  <header class="border-outline border-b bg-black/80 px-5 py-4">
    <div class="flex items-start justify-between gap-4">
      <div class="min-w-0">
        <p class="text-muted font-mono text-[11px] tracking-[0.3em] uppercase">Reading Pane</p>
        <h2 class="text-bright mt-2 truncate font-serif text-2xl">@{post.author_handle}</h2>
        <p class="text-muted mt-1 font-mono text-xs">{formatLongDateTime(post.created_at)}</p>
      </div>

      <button
        type="button"
        onclick={onClose}
        class="border-outline bg-surface text-muted hover:text-bright rounded-full border px-3 py-1.5 font-mono text-xs transition-colors">
        Close
      </button>
    </div>
  </header>

  <div class="flex-1 overflow-y-auto px-5 py-5">
    <div class="border-outline rounded-2xl border bg-black/60 p-5">
      <div class="mb-5 flex flex-wrap gap-2">
        <span
          class="bg-primary/15 text-primary rounded-full px-3 py-1 font-mono text-[11px] tracking-[0.18em] uppercase">
          {post.source}
        </span>
        <span
          class="border-outline text-muted flex items-center gap-1 rounded-full border px-3 py-1 font-mono text-[11px]">
          <i class="i-ri-heart-line"></i>
          <span>{post.like_count || 0}</span>
        </span>
        <span class="border-outline text-muted flex items-center rounded-full border px-3 py-1 font-mono text-[11px]">
          <i class="i-ri-repeat-line"></i>
          <span>{post.repost_count || 0}</span>
        </span>
        <span class="border-outline text-muted flex items-center rounded-full border px-3 py-1 font-mono text-[11px]">
          <i class="i-ri-message-2-line"></i>
          <span>{post.reply_count || 0}</span>
        </span>
      </div>

      <div class="space-y-4">
        <p class="text-muted font-mono text-xs tracking-[0.22em] uppercase">Post</p>
        <div class="text-bright font-mono text-sm leading-7 text-pretty">
          <PostText text={post.text} facetsJson={post.facets} />
        </div>
      </div>
    </div>

    <dl class="border-outline bg-surface/80 mt-5 grid gap-4 rounded-2xl border p-4">
      <div>
        <dt class="text-muted font-mono text-[11px] tracking-[0.2em] uppercase">Post URI</dt>
        <dd class="text-bright mt-1 font-mono text-xs break-all">{post.uri}</dd>
      </div>
      <div>
        <dt class="text-muted font-mono text-[11px] tracking-[0.2em] uppercase">Indexed</dt>
        <dd class="text-bright mt-1 font-mono text-xs">{formatLongDateTime(post.indexed_at)}</dd>
      </div>
    </dl>
  </div>

  <footer class="border-outline border-t bg-black/75 px-5 py-4">
    <button
      type="button"
      onclick={openInBrowser}
      class="border-outline bg-surface text-bright hover:bg-outline flex w-full items-center rounded-xl border px-4 py-3 font-sans text-sm transition-colors">
      <span>Open on BlueSky</span>
      <i class="i-ri-blue-sky-fill m-2"></i>
      <i class="i-ri-external-link-line"></i>
    </button>
  </footer>
</aside>
