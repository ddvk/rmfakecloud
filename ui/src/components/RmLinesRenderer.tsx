import { useEffect, useRef, useState } from "react"

interface GlobalRendererInterface {
    RmRenderer: () => Promise<RmRendererApi>,
}

interface RmRendererApi {
    callMain(args: string[]): Promise<void>,
    FS_writeFile(name: string, data: Uint8Array): void,
    FS_readFile(name:string): Uint8Array,
}

export interface DocumentEnvironment {
    rootDocId: string,
    pageCount: number,
    loadSubFile(name: string): Promise<Uint8Array>,
}

interface RmRendererProps {
    environment: DocumentEnvironment | null,
    page: number,
}

interface Page {
    uuid: string,
}

interface RMDocumentContent {
    cPages?: {
        pages?: {
            id: string,
            deleted?: {}
        }[]
    }
}

async function render(renderer: RmRendererApi, rawData: Uint8Array) {
    renderer.FS_writeFile("/rm", rawData);
    await renderer.callMain(["/rm", "/bmp"]);
    const bitmap = renderer.FS_readFile("/bmp");
    const dv = new DataView(bitmap.buffer);
    const rawContent = new Uint8ClampedArray(bitmap.subarray(8));
    const width = dv.getUint32(0);
    const height = dv.getUint32(4);
    const data = new ImageData(rawContent, dv.getUint32(0), dv.getUint32(4));
    return { width, height, data };
}

async function loadPages(env: DocumentEnvironment): Promise<Page[]> {
    const contentFile = `${env.rootDocId}.content`;
    const contentData = JSON.parse(
        new TextDecoder().decode(
            await env.loadSubFile(contentFile),
        )
    ) as RMDocumentContent;
    return (contentData.cPages?.pages || []).filter(e => !e.deleted).map(e => ({ uuid: e.id }));
}

export function RmLinesRenderer(props: RmRendererProps) {
    const [module, setModule] = useState<RmRendererApi | null>(null);
    const [ready, setReady] = useState(false);
    const [renderError, setRenderError] = useState(false);
    useEffect(() => {
        (async () => {
            setModule(await (window as unknown as GlobalRendererInterface).RmRenderer());
        })();
    }, []);

    // Load pages in a sane format
    const [pages, setPages] = useState<Page[]>([]);
    useEffect(() => {
        (async () => {
            setPages([]);
            if(props.environment)
                setPages(await loadPages(props.environment));
        })()
    }, [props.environment]);
    const canvasRef = useRef(null);
    const [context, setContext] = useState<CanvasRenderingContext2D | null>(null);
    const [drawnData, setDrawnData] = useState<{width: number, height: number, data: ImageData} | null>(null);
    useEffect(() => {
        if(canvasRef.current) {
            setContext(canvasRef.current.getContext('2d'));
        } else {
            setContext(null);
        }
    }, [canvasRef]);
    useEffect(() => {
        (async () => {
            if(props.environment && pages[props.page]) {
                setReady(false);
                setRenderError(false);
                console.log("Rendering...");
                try{
                    setDrawnData(await render(
                        module,
                        await props.environment.loadSubFile(
                            `${props.environment.rootDocId}/${pages[props.page].uuid}.rm`
                        )
                    ));
                }catch(ex) {
                    console.log(ex);
                    setRenderError(true);
                }
                setReady(true);
            }
        })();
    }, [module, props.page, pages, props.environment]);
    useEffect(() => {
        if(!context || !drawnData) return;
        context.canvas.width = drawnData.width;
        context.canvas.height = drawnData.height;
        context.putImageData(drawnData.data, 0, 0);
    }, [context, drawnData]);
    return <div className="rmrender-wrapper">
        {!ready && <span>Rendering...</span>}
        {renderError && <span>Render Error!</span>}
        <canvas ref={canvasRef} className="rmrender" style={{
            backgroundColor: 'white',
            display: (renderError || !ready) ? 'none' : undefined,
        }}></canvas>
    </div>
}
