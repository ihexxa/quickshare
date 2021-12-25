import * as React from "react";
import { List } from "immutable";

import { RiShareBoxLine } from "@react-icons/all-files/ri/RiShareBoxLine";
import { RiFolderSharedFill } from "@react-icons/all-files/ri/RiFolderSharedFill";
import { RiEmotionSadLine } from "@react-icons/all-files/ri/RiEmotionSadLine";

import { getErrMsg } from "../common/utils";
import { alertMsg, confirmMsg } from "../common/env";
import { updater } from "./state_updater";
import { ICoreState, MsgProps, UIProps } from "./core_state";
import { LoginProps } from "./pane_login";
import { Flexbox } from "./layout/flexbox";
import { Container } from "./layout/container";
import { Rows, Row } from "./layout/rows";

export interface SharingsProps {
  sharings: List<string>;
}

export interface Props {
  sharingsInfo: SharingsProps;
  msg: MsgProps;
  login: LoginProps;
  ui: UIProps;
  update?: (updater: (prevState: ICoreState) => ICoreState) => void;
}

export interface State {}

export class SharingsPanel extends React.Component<Props, State, {}> {
  constructor(p: Props) {
    super(p);
    this.state = {};
  }

  deleteSharing = async (dirPath: string) => {
    try {
      const deleteStatus = await updater().deleteSharing(dirPath);
      if (deleteStatus !== "") {
        throw deleteStatus;
      }
      updater().setSharing(false);

      await this.listSharings();
    } catch (e: any) {
      alertMsg(getErrMsg(this.props.msg.pkg, "op.fail", status));
    }
  };

  listSharings = async () => {
    const status = await updater().listSharings();
    if (status !== "") {
      alertMsg(getErrMsg(this.props.msg.pkg, "op.fail", status));
    }
    this.props.update(updater().updateFilesInfo);
    this.props.update(updater().updateSharingsInfo);
  };

  makeRows = (sharings: List<string>): List<Row> => {
    const sharingRows = sharings.map((dirPath: string) => {
      const sharingURL = `${
        document.location.href.split("?")[0]
      }?dir=${encodeURIComponent(dirPath)}`;

      const row1 = (
        <div>
          <div className="info">{dirPath}</div>

          <div className="op">
            <button
              onClick={() => {
                this.deleteSharing(dirPath);
              }}
            >
              {this.props.msg.pkg.get("browser.delete")}
            </button>
          </div>
        </div>
      );

      const elem = (
        <div className="sharing-item" key={dirPath}>
          {row1}
          <div className="desc">{sharingURL}</div>
          <div className="hr grey0-bg"></div>
        </div>
      );

      return {
        elem,
        sortVals: List([dirPath]),
        val: dirPath,
      };
    });

    return sharingRows;
  };

  updateSharings = (sharings: List<Object>) => {
    const newSharings = sharings as List<string>;
    updater().updateSharings(newSharings);
    this.props.update(updater().updateSharingsInfo);
  };

  render() {
    const sharingRows = this.makeRows(this.props.sharingsInfo.sharings);
    const view = (
      <Rows
        rows={sharingRows}
        sortKeys={List([this.props.msg.pkg.get("item.path")])}
        updateRows={this.updateSharings}
      />
    );
    const noSharingView = (
      <Container>
        <Flexbox
          children={List([
            <RiEmotionSadLine size="4rem" className="margin-r-m red0-font" />,
            <span>
              <h3 className="title-l">
                {this.props.msg.pkg.get("share.404.title")}
              </h3>
              <span className="desc-l grey0-font">
                {this.props.msg.pkg.get("share.404.desc")}
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
      this.props.sharingsInfo.sharings.size === 0 ? (
        noSharingView
      ) : (
        <Container>
          <Flexbox
            children={List([
              <Flexbox
                children={List([
                  <RiShareBoxLine
                    size="3rem"
                    className="margin-r-m black-font"
                  />,

                  <span>
                    <span className="title-m bold">
                      {this.props.msg.pkg.get("browser.share.title")}
                    </span>
                    <span className="desc-m grey0-font">
                      {this.props.msg.pkg.get("browser.share.desc")}
                    </span>
                  </span>,
                ])}
              />,

              <span></span>,
            ])}
            className="margin-b-l"
          />

          {view}
        </Container>
      );

    return <div id="sharing-list">{list}</div>;
  }
}
