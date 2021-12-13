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
import { getIcon } from "./visual/icons";

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
    return updater()
      .deleteSharing(dirPath)
      .then((status: string) => {
        if (status !== "") {
          alertMsg(getErrMsg(this.props.msg.pkg, "op.fail", status));
        } else {
          updater().setSharing(false);
          return this.listSharings();
        }
      });
  };

  listSharings = async () => {
    return updater()
      .listSharings()
      .then((status: string) => {
        if (status !== "") {
          alertMsg(getErrMsg(this.props.msg.pkg, "op.fail", status));
        }
        this.props.update(updater().updateFilesInfo);
        this.props.update(updater().updateSharingsInfo);
      });
  };

  render() {
    const nameWidthClass = `item-name item-name-${
      this.props.ui.isVertical ? "vertical" : "horizontal"
    } pointer`;

    const sharingList = this.props.sharingsInfo.sharings.map(
      (dirPath: string) => {
        return (
          <div id="share-list" key={dirPath}>
            <Flexbox
              children={List([
                <Flexbox
                  children={List([
                    <RiFolderSharedFill
                      size="3rem"
                      className="purple0-font margin-r-m"
                    />,
                    <span className={nameWidthClass}>{dirPath}</span>,
                  ])}
                />,

                <span
                  onClick={() => {
                    this.deleteSharing(dirPath);
                  }}
                  className="margin-r-m"
                >
                  {getIcon("RiDeleteBin2Fill", "1.8rem", "red0")}
                </span>,

                <span>
                  <input
                    type="text"
                    readOnly
                    value={`${
                      document.location.href.split("?")[0]
                    }?dir=${encodeURIComponent(dirPath)}`}
                  />
                </span>,
              ])}
              childrenStyles={List([
                { alignItems: "center" },
                { alignItems: "center", justifyContent: "flex-end" },
                { alignItems: "center", justifyContent: "flex-end" },
              ])}
            />
          </div>
        );
      }
    );

    const list =
      this.props.sharingsInfo.sharings.size === 0 ? (
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
          />

          {sharingList}
        </Container>
      );

    return <div id="sharing-list">{list}</div>;
  }
}
