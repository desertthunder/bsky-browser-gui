export type FacetFeature = { [key: string]: any; $type: string };

export type FacetByteSlice = { byteStart: number; byteEnd: number };

export interface Facet {
  index: FacetByteSlice;
  features: FacetFeature[];
}

type FacetKind = "link" | "mention" | "tag" | "text";

export type RenderedFacet = { type: FacetKind; text: string; href?: string; did?: string; tag?: string };

/**
 * Convert UTF-8 byte offsets to JS string indices (UTF-16 code units)
 */
function byteOffsetToCharIndex(text: string, byteOffset: number): number {
  const encoder = new TextEncoder();
  let currentByte = 0;

  for (const [i, char] of Array.from(text).entries()) {
    const charBytes = encoder.encode(char).length;

    if (currentByte >= byteOffset) {
      return i;
    }

    currentByte += charBytes;
  }

  return text.length;
}

/**
 * Parse a facets JSON string and return parsed Facet objects
 */
export function parseFacets(facetsJson: string): Facet[] {
  if (!facetsJson) return [];

  try {
    const parsed = JSON.parse(facetsJson);
    if (Array.isArray(parsed)) {
      return parsed as Facet[];
    }
  } catch (e) {
    console.warn("Failed to parse facets:", e);
  }

  return [];
}

/**
 * Render facets into an array of RenderedFacet objects
 * This converts byte offsets to JS string indices and extracts the text segments
 */
export function renderFacets(text: string, facets: Facet[]): RenderedFacet[] {
  if (!facets || facets.length === 0) {
    return [{ type: "text", text }];
  }

  const sortedFacets = [...facets].sort((a, b) => a.index.byteStart - b.index.byteStart);

  const result: RenderedFacet[] = [];
  let lastByteEnd = 0;

  for (const facet of sortedFacets) {
    if (facet.index.byteStart > lastByteEnd) {
      const beforeStart = byteOffsetToCharIndex(text, lastByteEnd);
      const beforeEnd = byteOffsetToCharIndex(text, facet.index.byteStart);
      const beforeText = text.slice(beforeStart, beforeEnd);
      if (beforeText) {
        result.push({ type: "text", text: beforeText });
      }
    }

    const facetStart = byteOffsetToCharIndex(text, facet.index.byteStart);
    const facetEnd = byteOffsetToCharIndex(text, facet.index.byteEnd);
    const facetText = text.slice(facetStart, facetEnd);
    let renderedFacet: RenderedFacet = { type: "text", text: facetText };

    for (const feature of facet.features) {
      const type = feature.$type;

      if (type === "app.bsky.richtext.facet#link") {
        renderedFacet = { type: "link", text: facetText, href: feature.uri };
        break;
      }

      if (type === "app.bsky.richtext.facet#mention") {
        renderedFacet = {
          type: "mention",
          text: facetText,
          did: feature.did,
          href: `https://bsky.app/profile/${feature.did}`,
        };
        break;
      }

      if (type === "app.bsky.richtext.facet#tag") {
        renderedFacet = {
          type: "tag",
          text: facetText,
          tag: feature.tag,
          href: `https://bsky.app/search?q=%23${encodeURIComponent(feature.tag)}`,
        };
        break;
      }
    }

    result.push(renderedFacet);

    lastByteEnd = facet.index.byteEnd;
  }

  const encoder = new TextEncoder();
  const textBytes = encoder.encode(text).length;
  if (lastByteEnd < textBytes) {
    const remainingStart = byteOffsetToCharIndex(text, lastByteEnd);
    const remainingText = text.slice(remainingStart);
    if (remainingText) {
      result.push({ type: "text", text: remainingText });
    }
  }

  return result;
}

/**
 * Get plain text with facets stripped (for truncation)
 */
export function getPlainText(text: string, facets: Facet[]): string {
  return text;
}

/**
 * Truncate rendered facets to a maximum length while preserving facet boundaries
 */
export function truncateRenderedFacets(
  rendered: RenderedFacet[],
  maxLen: number,
): { facets: RenderedFacet[]; truncated: boolean } {
  let currentLength = 0;
  const result: RenderedFacet[] = [];
  let truncated = false;

  for (const facet of rendered) {
    const remaining = maxLen - currentLength;

    if (remaining <= 0) {
      truncated = true;
      break;
    }

    if (facet.text.length <= remaining) {
      result.push(facet);
      currentLength += facet.text.length;
    } else {
      const truncatedText = facet.text.slice(0, remaining) + "...";
      result.push({ ...facet, text: truncatedText });
      truncated = true;
      break;
    }
  }

  return { facets: result, truncated };
}
