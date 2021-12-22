import * as React from "react";
import { List } from "immutable";
import FileSize from "filesize";

import { RiUploadCloudFill } from "@react-icons/all-files/ri/RiUploadCloudFill";
import { RiUploadCloudLine } from "@react-icons/all-files/ri/RiUploadCloudLine";
import { RiEmotionSadLine } from "@react-icons/all-files/ri/RiEmotionSadLine";

import { alertMsg } from "../common/env";
import { getErrMsg } from "../common/utils";
import { updater } from "./state_updater";
import { ICoreState, MsgProps, UIProps } from "./core_state";
import { LoginProps } from "./pane_login";
import { UploadEntry, UploadState } from "../worker/interface";
import { Flexbox } from "./layout/flexbox";
import { Container } from "./layout/container";
import { Rows, Row } from "./layout/rows";

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

  deleteUpload = async (filePath: string): Promise<void> => {
    try {
      const deleteStatus = await updater().deleteUpload(filePath);
      if (deleteStatus !== "") {
        throw deleteStatus;
      }

      const statuses = await Promise.all([
        updater().refreshUploadings(),
        updater().self(),
      ]);
      if (statuses.join("") !== "") {
        throw statuses.join(";");
      }

      this.props.update(updater().updateLogin);
      this.props.update(updater().updateUploadingsInfo);
    } catch (status: any) {
      alertMsg(getErrMsg(this.props.msg.pkg, "op.fail", status.toString()));
    }
  };

  stopUploading = (filePath: string) => {
    updater().stopUploading(filePath);
    this.props.update(updater().updateUploadingsInfo);
  };

  makeRowsInputs = (uploadings: List<UploadEntry>): List<Row> => {
    const uploadingRows = uploadings.map((uploading: UploadEntry) => {
      const pathParts = uploading.filePath.split("/");
      const fileName = pathParts[pathParts.length - 1];
      const progress = Math.floor((uploading.uploaded / uploading.size) * 100);

      // const title = <Flexbox children={List([])} />;

      const op = (
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
        </div>
      );

      const progressBar = (
        <div className="progress-grey">
          <div
            className="progress-green"
            style={{ width: `${progress}%` }}
          ></div>
        </div>
      );

      const errorInfo =
        uploading.err.trim() === "" ? null : (
          <div className="error">{uploading.err.trim()}</div>
        );

      const elem = (
        <div key={uploading.filePath} className="upload-item">
          <Flexbox
            children={List([
              <RiUploadCloudLine
                size="3rem"
                id="icon-upload"
                className="margin-r-m blue0-font"
              />,

              <div className={`font-s`}>
                <span className="">{fileName}&nbsp;</span>
                <span className="desc grey0-font">
                  {FileSize(uploading.uploaded, { round: 0 })}
                  &nbsp;/&nbsp;
                  {FileSize(uploading.size, {
                    round: 0,
                  })}
                  &nbsp;/&nbsp;
                  {`${progress}%`}
                </span>
              </div>,

              op,
            ])}
            childrenStyles={List([
              { flex: "0 0 auto" },
              {},
              { justifyContent: "flex-end" },
            ])}
          />
          {progressBar}
          {errorInfo}
        </div>
      );

      // file path, size, progress
      const sortVals = List<string>([
        uploading.filePath,
        `${uploading.size}`,
        `${progress}`,
      ]);
      return {
        elem,
        sortVals,
        val: uploading,
      };
    });

    return uploadingRows;
  };

  updateUploadings = (uploadings: Object) => {
    const newUploadings = uploadings as List<UploadEntry>;
    updater().updateUploadings(newUploadings);
    this.props.update(updater().updateUploadingsInfo);
  };

  render() {
    const uploadingRows = this.makeRowsInputs(
      this.props.uploadingsInfo.uploadings
    );
    const sortKeys = List([
      this.props.msg.pkg.get("item.path"),
      this.props.msg.pkg.get("item.size"),
      this.props.msg.pkg.get("item.progress"),
    ]);
    const view = (
      <Rows
        sortKeys={sortKeys}
        rows={uploadingRows}
        updateRows={this.updateUploadings}
      />
    );

    const noUploadingView = (
      <Container>
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
        />
      </Container>
    );

    const list =
      this.props.uploadingsInfo.uploadings.size === 0 ? (
        noUploadingView
      ) : (
        <Container>
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

          {view}
        </Container>
      );

    return <div id="upload-list">{list}</div>;
  }
}
