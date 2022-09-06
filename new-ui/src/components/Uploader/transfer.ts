import PQueue from 'p-queue'

import { uploadDocument, UploadedFile } from '../../api/document'
import { v4 as uuidv4 } from 'uuid'

export default class Transfer<T extends File> {
  concurrent: number = 1
  deduplicated: boolean = false

  uploadedFiles: UploadedFile[] = []
  queue: PQueue

  onDuplicatedFiles?: (files: T[]) => void
  onChange?: (file: UploadedFile, files: UploadedFile[]) => void

  constructor(concurrent: number, deduplicated: boolean) {
    this.concurrent = concurrent
    this.deduplicated = deduplicated

    this.queue = new PQueue({
      concurrency: this.concurrent,
      timeout: 5 * 60 * 1000 // 5mins
    })
  }

  uploadFiles(files: T[], forced: boolean = false) {
    let duplicatedFiles: T[] = []
    if (this.deduplicated && !forced) {
      const [pendingFiles, duplicated] = this.checkDuplicatedFiles(...files)
      files = pendingFiles
      duplicatedFiles = duplicated
    }

    this.initUploadedFiles(files)
    this.uploadedFiles.forEach((f) => {
      if (f.status === 'pending') {
        this.uploadFile(f)
      }
    })

    if (duplicatedFiles.length > 0 && this.onDuplicatedFiles) {
      this.onDuplicatedFiles(duplicatedFiles)
    }
  }

  protected initUploadedFiles(files: T[]) {
    files.forEach((file) => {
      const ctrl = new AbortController()
      let uploadedFile: UploadedFile = {
        id: uuidv4().toString(),
        uploadedSize: 0,
        status: 'pending',
        cancel: ctrl,
        file: file
      }

      this.uploadedFiles.push(uploadedFile)
    })
  }

  protected async uploadFile(file: UploadedFile) {
    await this.queue.add(async () => {
      return await uploadDocument(file, (f) => {
        this.onChange && this.onChange(f, this.uploadedFiles)
      })
    })
  }

  protected checkDuplicatedFiles(...files: T[]): [T[], T[]] {
    let duplicatedFiles: T[] = []
    let pendingFiles: T[] = []

    files.forEach((f) => {
      const i = this.uploadedFiles.findIndex(
        (enqueueFile) =>
          enqueueFile.file.name == f.name && enqueueFile.file.lastModified == f.lastModified
      )

      if (i === -1) {
        pendingFiles.push(f)
      } else {
        duplicatedFiles.push(f)
      }
    })

    return [pendingFiles, duplicatedFiles]
  }
}
