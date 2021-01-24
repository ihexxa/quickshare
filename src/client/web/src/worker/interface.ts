import { Map } from "immutable";

export interface UploadEntry {
  file: File;
  filePath: string;
  size: number;
  uploaded: number;
  runnable: boolean;
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
  infos: Map<string, UploadEntry>;
}

export type FileWorkerReq = SyncReq;

export type ErrKind = "worker.resp.err";
export const errKind: ErrKind = "worker.resp.err";
export interface ErrResp extends WorkerEvent {
  kind: ErrKind;
  err: string;
}

// caller should combine uploaded and done to see if the upload is successfully finished
export type UploadInfoKind = "worker.resp.info";
export const uploadInfoKind: UploadInfoKind = "worker.resp.info";
export interface UploadInfoResp extends WorkerEvent {
  kind: UploadInfoKind;
  filePath: string;
  uploaded: number;
  runnable: boolean;
  err: string;
}

export type FileWorkerResp = ErrResp | UploadInfoResp;

export class MockWorker {
  constructor() {}
  onmessage = (event: MessageEvent): void => {};
  postMessage = (event: FileWorkerReq): void => {};
}