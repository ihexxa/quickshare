import * as React from "react";
import { List } from "immutable";

import { RiShareBoxLine } from "@react-icons/all-files/ri/RiShareBoxLine";
import { RiFolderSharedFill } from "@react-icons/all-files/ri/RiFolderSharedFill";
import { RiEmotionSadLine } from "@react-icons/all-files/ri/RiEmotionSadLine";

import { alertMsg, confirmMsg } from "../common/env";
import { updater } from "./state_updater";
import { ICoreState, MsgProps, UIProps } from "./core_state";
import { LoginProps } from "./pane_login";
import { Flexbox } from "./layout/flexbox";
import { Container } from "./layout/container";

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

  addSharing = async () => {
    return updater()
      .addSharing()
      .then((ok) => {
        if (!ok) {
          alertMsg(this.props.msg.pkg.get("browser.share.add.fail"));
        } else {
          updater().setSharing(true);
          return this.listSharings();
        }
      })
      .then(() => {
        this.props.update(updater().updateFilesInfo);
        this.props.update(updater().updateSharingsInfo);
      });
  };

  deleteSharing = async (dirPath: string) => {
    return updater()
      .deleteSharing(dirPath)
      .then((ok) => {
        if (!ok) {
          alertMsg(this.props.msg.pkg.get("browser.share.del.fail"));
        } else {
          updater().setSharing(false);
          return this.listSharings();
        }
      })
      .then(() => {
        this.props.update(updater().updateFilesInfo);
        this.props.update(updater().updateFilesInfo);
      });
  };

  listSharings = async () => {
    return updater()
      .listSharings()
      .then((ok) => {
        if (ok) {
          this.props.update(updater().updateFilesInfo);
          this.props.update(updater().updateSharingsInfo);
        }
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
                    <span>{dirPath}</span>,
                  ])}
                />,

                <span>
                  <input
                    type="text"
                    readOnly
                    className="float-input"
                    value={`${
                      document.location.href.split("?")[0]
                    }?dir=${encodeURIComponent(dirPath)}`}
                  />
                  <button
                    onClick={() => {
                      this.deleteSharing(dirPath);
                    }}
                    className="float-input"
                  >
                    {this.props.msg.pkg.get("browser.share.del")}
                  </button>
                </span>,
              ])}
              childrenStyles={List([{}, { justifyContent: "flex-end" }])}
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
