<script lang="ts">
  import { parseFacets, renderFacets, truncateRenderedFacets, type RenderedFacet } from "../facets";

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

  function getFacetLabel(facet: RenderedFacet) {
    switch (facet.type) {
      case "link":
        return `Open link ${facet.href ?? facet.text}`;
      case "mention":
        return `Open profile ${facet.text}`;
      case "tag":
        return `Search tag ${facet.tag ?? facet.text}`;
      default:
        return facet.text;
    }
  }
</script>

<span class={className}>
  {#each renderedFacets as facet}
    {#if facet.type === "link"}
      <a
        href={facet.href}
        target="_blank"
        rel="noopener noreferrer"
        title={getFacetLabel(facet)}
        aria-label={getFacetLabel(facet)}
        class="text-primary hover:text-primary-bright hover:underline"
        onclick={(e) => e.stopPropagation()}>
        {facet.text}
      </a>
    {:else if facet.type === "mention"}
      <a
        href={facet.href}
        target="_blank"
        rel="noopener noreferrer"
        title={getFacetLabel(facet)}
        aria-label={getFacetLabel(facet)}
        class="text-primary hover:text-primary-bright hover:underline"
        onclick={(e) => e.stopPropagation()}>
        {facet.text}
      </a>
    {:else if facet.type === "tag"}
      <a
        href={facet.href}
        target="_blank"
        rel="noopener noreferrer"
        title={getFacetLabel(facet)}
        aria-label={getFacetLabel(facet)}
        class="text-secondary hover:text-secondary-bright hover:underline"
        onclick={(e) => e.stopPropagation()}>
        {facet.text}
      </a>
    {:else}
      <span class="font-sans">{facet.text}</span>
    {/if}
  {/each}
</span>
