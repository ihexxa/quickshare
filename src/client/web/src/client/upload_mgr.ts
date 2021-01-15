import { Map } from "immutable";

import { UploadInfo } from "./";
import { FileUploader } from "./uploader";

export class UploadMgr {
  private static infos = Map<string, UploadInfo>();
  private static uploaders = Map<string, FileUploader>();
  private static statusCb: (infos: Map<string, UploadInfo>) => void;

  static setStatusCb = (cb: (infos: Map<string, UploadInfo>) => void) => {
    UploadMgr.statusCb = cb;
  };

  static list = (): Map<string, UploadInfo> => {
    return UploadMgr.infos;
  };

  static start = (file: File, filePath: string) => {
    const uploader = new FileUploader(file, filePath, UploadMgr.progressCb);
    UploadMgr.uploaders = UploadMgr.uploaders.set(filePath, uploader);
    UploadMgr.infos = UploadMgr.infos.set(filePath, {
      realFilePath: filePath,
      size: file.size,
      uploaded: 0,
    });
    return uploader.start();
  };

  static delete = (filePath: string) => {
    const uploader = UploadMgr.uploaders.get(filePath);
    if (uploader != null) {
      uploader.stop();
      UploadMgr.uploaders = UploadMgr.uploaders.delete(filePath);
      UploadMgr.infos = UploadMgr.infos.delete(filePath);
    }
  };

  static progressCb = (filePath: string, uploaded: number) => {
    const info = UploadMgr.infos.get(filePath);

    if (info != null) {
      if (info.size === uploaded) {
        UploadMgr.infos = UploadMgr.infos.delete(filePath);
        UploadMgr.uploaders = UploadMgr.uploaders.delete(filePath);
        UploadMgr.statusCb(UploadMgr.infos);
      } else {
        UploadMgr.infos = UploadMgr.infos.set(filePath, {
          ...info,
          uploaded: uploaded,
        });
        UploadMgr.statusCb(UploadMgr.infos);
      }
    }
  };
}
