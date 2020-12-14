import axios, { AxiosRequestConfig, AxiosResponse } from "axios";

import {MetadataResp, UploadStatusResp, ListResp} from "./files";

export class FilesClient {
  private url: string;
  private createMockResp: number;
  private deleteMockResp: number;
  private metadataMockResp: MetadataResp | null;
  private mkdirMockResp: number | null;
  private moveMockResp: number;
  private uploadChunkMockResp: UploadStatusResp | null;
  private uploadStatusMockResp: UploadStatusResp | null;
  private listMockResp: ListResp | null;

  constructor(url: string) {
    this.url = url;
  }

  createMock = (resp: number) => {
    this.createMockResp = resp;
  }
  deleteMock = (resp: number) => {
    this.deleteMockResp = resp;
  }
  metadataMock = (resp: MetadataResp | null) => {
    this.metadataMockResp = resp;
  }
  mkdirMock = (resp: number | null) => {
    this.mkdirMockResp = resp;
  }
  moveMock = (resp: number) => {
    this.moveMockResp = resp;
  }
  uploadChunkMock = (resp: UploadStatusResp | null) => {
    this.uploadChunkMockResp = resp;
  }

  uploadStatusMock = (resp: UploadStatusResp | null) => {
    this.uploadStatusMockResp = resp;
  }

  listMock = (resp: ListResp | null) => {
    this.listMockResp = resp;
  }

  async create(filePath: string, fileSize: number): Promise<number> {
    return this.createMockResp;
  }

  async delete(filePath: string): Promise<number> {
    return this.deleteMockResp;
  }

  async metadata(filePath: string): Promise<MetadataResp | null> {
    return this.metadataMockResp;
  }

  async mkdir(dirpath: string): Promise<number | null> {
    return this.mkdirMockResp;
  }

  async move(oldPath: string, newPath: string): Promise<number> {
    return this.moveMockResp;
  }

  async uploadChunk(
    filePath: string,
    content: string | ArrayBuffer,
    offset: number
  ): Promise<UploadStatusResp | null> {
    return this.uploadChunkMockResp;
  }

  async uploadStatus(filePath: string): Promise<UploadStatusResp | null> {
    return this.uploadStatusMockResp;
  }

  async list(dirPath: string): Promise<ListResp | null> {
    return this.listMockResp;
  }
}
