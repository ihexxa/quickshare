declare module "worker-loader!*" {
  class UploadWorker extends Worker {
    constructor();
  }

  export = UploadWorker;
}
