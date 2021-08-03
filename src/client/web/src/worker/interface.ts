export const enum UploadState {
  Created,
  Ready,
  Uploading,
  Stopped,
  Error,
}

export interface UploadStatus {
  filePath: string;
  uploaded: number;
  state: UploadState;
  err: string;
}

export interface UploadEntry {
  file: File;
  filePath: string;
  size: number;
  uploaded: number;
  state: UploadState;
  err: string;
}

export type eventKind = SyncReqKind | ErrKind | UploadInfoKind;
export interface WorkerEvent {
  kind: eventKind;
}

export type SyncReqKind = "worker.req.sync";
export const syncReqKind: SyncReqKind = "worker.req.sync";

export interface SyncReq extends WorkerEvent {
  kind: SyncReqKind;
  file: File,
  filePath: string;
  size: number;
  uploaded: number;
  created: boolean;
}

export type FileWorkerReq = SyncReq;

export type ErrKind = "worker.resp.err";
export const errKind: ErrKind = "worker.resp.err";
export interface ErrResp extends WorkerEvent {
  kind: ErrKind;
  filePath: string;
  err: string;
}

// caller should combine uploaded and done to see if the upload is successfully finished
export type UploadInfoKind = "worker.resp.info";
export const uploadInfoKind: UploadInfoKind = "worker.resp.info";
export interface UploadInfoResp extends WorkerEvent {
  kind: UploadInfoKind;
  filePath: string;
  uploaded: number;
  state: UploadState;
  err: string;
}

export type FileWorkerResp = ErrResp | UploadInfoResp;

export class MockWorker {
  constructor() {}
  onmessage = (event: MessageEvent): void => {};
  postMessage = (event: FileWorkerReq): void => {};
}