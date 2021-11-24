import * as React from "react";
import { List } from "immutable";
import FileSize from "filesize";

import { RiUploadCloudFill } from "@react-icons/all-files/ri/RiUploadCloudFill";
import { RiUploadCloudLine } from "@react-icons/all-files/ri/RiUploadCloudLine";
import { RiEmotionSadLine } from "@react-icons/all-files/ri/RiEmotionSadLine";

import { alertMsg } from "../common/env";
import { updater } from "./state_updater";
import { ICoreState, MsgProps, UIProps } from "./core_state";
import { LoginProps } from "./pane_login";
import { UploadEntry, UploadState } from "../worker/interface";
import { Flexbox } from "./layout/flexbox";

export interface UploadingsProps {
  uploadings: List<UploadEntry>;
  uploadFiles: List<File>;
}

export interface Props {
  uploadingsInfo: UploadingsProps;
  msg: MsgProps;
  login: LoginProps;
  ui: UIProps;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export interface State {}

export class UploadingsPanel extends React.Component<Props, State, {}> {
  constructor(p: Props) {
    super(p);
    this.state = {};
  }

  addUploads = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.files.length > 1000) {
      alertMsg(this.props.msg.pkg.get("err.tooManyUploads"));
      return;
    }

    let fileList = List<File>();
    for (let i = 0; i < event.target.files.length; i++) {
      fileList = fileList.push(event.target.files[i]);
    }
    updater().addUploads(fileList);
    this.props.update(updater().updateUploadingsInfo);
  };

  deleteUpload = (filePath: string): Promise<void> => {
    return updater()
      .deleteUpload(filePath)
      .then((ok: boolean) => {
        if (!ok) {
          alertMsg(this.props.msg.pkg.get("browser.upload.del.fail"));
        }
        return updater().refreshUploadings();
      })
      .then(() => {
        return updater().self();
      })
      .then(() => {
        this.props.update(updater().updateUploadingsInfo);
        this.props.update(updater().updateLogin);
      });
  };

  stopUploading = (filePath: string) => {
    updater().stopUploading(filePath);
    this.props.update(updater().updateUploadingsInfo);
  };

  render() {
    const nameWidthClass = `item-name item-name-${
      this.props.ui.isVertical ? "vertical" : "horizontal"
    } pointer`;

    const uploadingList = this.props.uploadingsInfo.uploadings.map(
      (uploading: UploadEntry) => {
        const pathParts = uploading.filePath.split("/");
        const fileName = pathParts[pathParts.length - 1];

        return (
          <div key={uploading.filePath}>
            <Flexbox
              children={List([
                <span className="padding-m">
                  <Flexbox
                    children={List([
                      <RiUploadCloudLine
                        size="3rem"
                        id="icon-upload"
                        className="margin-r-m blue0-font"
                      />,

                      <div className={`${nameWidthClass}`}>
                        <span className="title-m">{fileName}</span>
                        <div className="desc-m grey0-font">
                          {FileSize(uploading.uploaded, { round: 0 })}
                          &nbsp;/&nbsp;{FileSize(uploading.size, { round: 0 })}
                        </div>
                      </div>,
                    ])}
                  />
                </span>,

                <div className="item-op">
                  <button
                    onClick={() => this.stopUploading(uploading.filePath)}
                    className="float-input"
                  >
                    {this.props.msg.pkg.get("browser.stop")}
                  </button>
                  <button
                    onClick={() => this.deleteUpload(uploading.filePath)}
                    className="float-input"
                  >
                    {this.props.msg.pkg.get("browser.delete")}
                  </button>
                </div>,
              ])}
              childrenStyles={List([{}, { justifyContent: "flex-end" }])}
            />
            {uploading.err.trim() === "" ? null : (
              <div className="error">{uploading.err.trim()}</div>
            )}
          </div>
        );
      }
    );

    return this.props.uploadingsInfo.uploadings.size === 0 ? (
      <div id="upload-list" className="container">
        <Flexbox
          children={List([
            <RiEmotionSadLine size="4rem" className="margin-r-m red0-font" />,
            <span>
              <h3 className="title-l">
                {this.props.msg.pkg.get("upload.404.title")}
              </h3>
              <span className="desc-l grey0-font">
                {this.props.msg.pkg.get("upload.404.desc")}
              </span>
            </span>,
          ])}
          childrenStyles={List([
            { flex: "auto", justifyContent: "flex-end" },
            { flex: "auto" },
          ])}
          className="padding-l"
        />
      </div>
    ) : (
      <div id="upload-list" className="container">
        <Flexbox
          children={List([
            <span className="upload-item">
              <Flexbox
                children={List([
                  <RiUploadCloudFill
                    size="3rem"
                    className="margin-r-m black-font"
                  />,

                  <span>
                    <span className="title-m bold">
                      {this.props.msg.pkg.get("browser.upload.title")}
                    </span>
                    <span className="desc-m grey0-font">
                      {this.props.msg.pkg.get("browser.upload.desc")}
                    </span>
                  </span>,
                ])}
              />
            </span>,

            <span></span>,
          ])}
        />

        {uploadingList}
      </div>
    );
  }
}
