import { useEffect, useRef, useState } from "react";

export interface DocumentEnvironment {
  rootDocId: string;
  pageCount: number;
  loadSubFile(name: string): Promise<Uint8Array>;
}

interface RmRendererApi {
  callMain(args: string[]): Promise<void>;
  FS_writeFile(name: string, data: Uint8Array): void;
  FS_readFile(name: string): Uint8Array;
}

interface RmLinesRendererProps {
  environment: DocumentEnvironment | null;
  page: number;
}

interface Page {
  uuid: string;
}

/** reMarkable .content uses top-level `pages` (UUID strings) and sometimes cPages.pages with { id }. */
function pageIdsFromContentJson(contentData: unknown): string[] {
  if (!contentData || typeof contentData !== "object") {
    return [];
  }
  const o = contentData as Record<string, unknown>;
  const cPages = o.cPages as { pages?: { id?: string; deleted?: unknown }[] } | undefined;
  if (cPages?.pages?.length) {
    return cPages.pages.filter((e) => e && !e.deleted && e.id).map((e) => String(e.id));
  }
  const pages = o.pages;
  if (!Array.isArray(pages)) {
    return [];
  }
  const out: string[] = [];
  for (const p of pages) {
    if (typeof p === "string" && p) {
      out.push(p);
    } else if (p && typeof p === "object" && "id" in p && (p as { id?: string }).id) {
      out.push(String((p as { id: string }).id));
    }
  }
  return out;
}

async function loadPages(env: DocumentEnvironment): Promise<Page[]> {
  const contentFile = `${env.rootDocId}.content`;
  const raw = await env.loadSubFile(contentFile);
  const contentData = JSON.parse(new TextDecoder().decode(raw)) as unknown;
  return pageIdsFromContentJson(contentData).map((uuid) => ({ uuid }));
}

async function loadRmBytes(env: DocumentEnvironment, uuid: string): Promise<Uint8Array> {
  const candidates = [`${uuid}.rm`, `${env.rootDocId}/${uuid}.rm`];
  let lastErr: Error | null = null;
  for (const path of candidates) {
    try {
      return await env.loadSubFile(path);
    } catch (e) {
      lastErr = e instanceof Error ? e : new Error(String(e));
    }
  }
  throw lastErr ?? new Error("no .rm blob for page");
}

async function renderFrame(renderer: RmRendererApi, rawData: Uint8Array) {
  renderer.FS_writeFile("/rm", rawData);
  await renderer.callMain(["/rm", "/bmp"]);
  const bitmap = renderer.FS_readFile("/bmp");
  const u8 = bitmap instanceof Uint8Array ? bitmap : new Uint8Array(bitmap);
  const dv = new DataView(u8.buffer, u8.byteOffset, u8.byteLength);
  const width = dv.getUint32(0, false);
  const height = dv.getUint32(4, false);
  const expected = 8 + width * height * 4;
  if (u8.byteLength < expected) {
    throw new Error(`bitmap too small: ${u8.byteLength} < ${expected}`);
  }
  const rawContent = new Uint8ClampedArray(u8.buffer, u8.byteOffset + 8, width * height * 4);
  const data = new ImageData(rawContent, width, height);
  return { width, height, data };
}

export function RmLinesRenderer(props: RmLinesRendererProps) {
  const [module, setModule] = useState<RmRendererApi | null>(null);
  const [moduleError, setModuleError] = useState(false);
  const [pages, setPages] = useState<Page[]>([]);
  const [pagesError, setPagesError] = useState(false);
  const [busy, setBusy] = useState(false);
  const [renderError, setRenderError] = useState(false);
  const [drawnData, setDrawnData] = useState<{
    width: number;
    height: number;
    data: ImageData;
  } | null>(null);
  const canvasRef = useRef<HTMLCanvasElement | null>(null);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const Rm = (window as unknown as { RmRenderer?: () => Promise<RmRendererApi> }).RmRenderer;
        if (!Rm) {
          setModuleError(true);
          return;
        }
        const m = await Rm();
        if (!cancelled) {
          setModule(m);
          setModuleError(false);
        }
      } catch {
        if (!cancelled) {
          setModuleError(true);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    (async () => {
      setPages([]);
      setPagesError(false);
      if (!props.environment) {
        return;
      }
      try {
        setPages(await loadPages(props.environment));
      } catch {
        setPagesError(true);
      }
    })();
  }, [props.environment]);

  useEffect(() => {
    if (!module || !props.environment) {
      setDrawnData(null);
      setBusy(false);
      setRenderError(false);
      return;
    }
    const pg = pages[props.page];
    if (!pg) {
      setDrawnData(null);
      setBusy(false);
      setRenderError(false);
      return;
    }

    let cancelled = false;
    (async () => {
      setBusy(true);
      setRenderError(false);
      setDrawnData(null);
      try {
        const raw = await loadRmBytes(props.environment!, pg.uuid);
        if (cancelled) {
          return;
        }
        const frame = await renderFrame(module, raw);
        if (!cancelled) {
          setDrawnData(frame);
        }
      } catch (e) {
        console.warn("RmLines render:", e);
        if (!cancelled) {
          setRenderError(true);
        }
      } finally {
        if (!cancelled) {
          setBusy(false);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [module, props.page, pages, props.environment]);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas || !drawnData) {
      return;
    }
    const ctx = canvas.getContext("2d");
    if (!ctx) {
      return;
    }
    canvas.width = drawnData.width;
    canvas.height = drawnData.height;
    ctx.putImageData(drawnData.data, 0, 0);
  }, [drawnData]);

  if (moduleError) {
    return (
      <p className="text-danger small" style={{ padding: 8 }}>
        Notebook line renderer (WASM) failed to load. Try refreshing the page.
      </p>
    );
  }

  if (!module) {
    return <span className="text-muted">Loading line renderer…</span>;
  }

  if (pagesError) {
    return (
      <p className="text-danger small" style={{ padding: 8 }}>
        Could not read notebook page list (.content).
      </p>
    );
  }

  if (pages.length === 0 && props.environment) {
    return <p className="text-muted small" style={{ padding: 8 }}>No pages in this notebook.</p>;
  }

  return (
    <div className="rmrender-wrapper">
      {busy && <span className="text-muted me-2">Rendering…</span>}
      {renderError && (
        <p className="text-danger small" style={{ padding: 8 }}>
          Could not render this page with the line engine (unsupported or corrupt .rm).
        </p>
      )}
      <canvas
        ref={canvasRef}
        className="rmrender"
        style={{
          backgroundColor: "white",
          display: renderError || !drawnData ? "none" : undefined,
          maxWidth: "100%",
          height: "auto",
        }}
      />
    </div>
  );
}
