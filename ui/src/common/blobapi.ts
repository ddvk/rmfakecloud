import { DocumentEnvironment } from "../components/RmLinesRenderer";

export async function generateDocumentEnvironment(id: string): Promise<DocumentEnvironment> {
    const blobTree = JSON.parse(
        new TextDecoder().decode(
            await (await fetch(`/ui/api/documents/${id}/blobs`)).arrayBuffer()
        )
    ) as {[path: string]: string};
    return {
        pageCount: Object.keys(blobTree).filter(e => e.endsWith('.rm')).length,
        rootDocId: id,
        loadSubFile: async (subFile) => {
            return new Uint8Array(
                await ((await fetch(`/ui/api/blobs/${blobTree[subFile]}`)).arrayBuffer())
            );
        }
    }
}
