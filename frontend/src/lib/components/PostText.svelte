<script lang="ts">
  import { parseFacets, renderFacets, truncateRenderedFacets } from "../facets";

  type Props = { text: string; facetsJson?: string; maxLength?: number; class?: string };

  let { text, facetsJson = "", maxLength = 0, class: className = "" }: Props = $props();

  let renderedFacets = $derived.by(() => {
    const facets = parseFacets(facetsJson);
    const rendered = renderFacets(text, facets);

    if (maxLength > 0 && maxLength < text.length) {
      const { facets: truncated } = truncateRenderedFacets(rendered, maxLength);
      return truncated;
    }

    return rendered;
  });
</script>

<span class={className}>
  {#each renderedFacets as facet}
    {#if facet.type === "link"}
      <a
        href={facet.href}
        target="_blank"
        rel="noopener noreferrer"
        class="text-primary hover:text-primary-bright hover:underline"
        onclick={(e) => e.stopPropagation()}>
        {facet.text}
      </a>
    {:else if facet.type === "mention"}
      <a
        href={facet.href}
        target="_blank"
        rel="noopener noreferrer"
        class="text-primary hover:text-primary-bright hover:underline"
        onclick={(e) => e.stopPropagation()}>
        {facet.text}
      </a>
    {:else if facet.type === "tag"}
      <a
        href={facet.href}
        target="_blank"
        rel="noopener noreferrer"
        class="text-secondary hover:text-secondary-bright hover:underline"
        onclick={(e) => e.stopPropagation()}>
        {facet.text}
      </a>
    {:else}
      <span class="font-sans">{facet.text}</span>
    {/if}
  {/each}
</span>
