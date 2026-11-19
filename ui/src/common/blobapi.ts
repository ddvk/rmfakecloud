import type { DocumentEnvironment } from "../components/RmLinesRenderer";
import constants from "./constants";

/**
 * Build a document environment for client-side .rm rendering (librm_lines WASM),
 * using the blob tree from sync 1.5+ storage. Returns null if the blob API is
 * unavailable (e.g. sync 1.0) or the document has no usable .content / .rm entries.
 */
export async function tryDocumentEnvironment(id: string): Promise<DocumentEnvironment | null> {
  let res: Response;
  try {
    res = await fetch(`${constants.ROOT_URL}/documents/${id}/blobs`, {
      credentials: "same-origin",
    });
  } catch {
    return null;
  }
  if (!res.ok) {
    return null;
  }
  let blobTree: Record<string, string>;
  try {
    blobTree = JSON.parse(new TextDecoder().decode(await res.arrayBuffer())) as Record<string, string>;
  } catch {
    return null;
  }
  const contentKey = `${id}.content`;
  if (!blobTree[contentKey]) {
    return null;
  }
  const hasRm = Object.keys(blobTree).some((k) => k.endsWith(".rm"));
  if (!hasRm) {
    return null;
  }

  return {
    rootDocId: id,
    pageCount: Object.keys(blobTree).filter((e) => e.endsWith(".rm")).length,
    loadSubFile: async (subFile: string) => {
      const hash = blobTree[subFile];
      if (!hash) {
        throw new Error(`missing blob mapping for ${subFile}`);
      }
      const br = await fetch(`${constants.ROOT_URL}/blobs/${hash}`, {
        credentials: "same-origin",
      });
      if (!br.ok) {
        throw new Error(`blob fetch failed ${br.status}`);
      }
      return new Uint8Array(await br.arrayBuffer());
    },
  };
}
