export interface FileWorkerReq {
  file: File;
  filePath: string;
  stop: boolean;
}

// caller should combine uploaded and done to see if the upload is successfully finished
export interface FileWorkerResp {
  filePath: string;
  uploaded: number;
  done: boolean;
  err?: string;
}
