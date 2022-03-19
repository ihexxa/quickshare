import * as React from "react";
import { List, Map } from "immutable";

import { BtnList } from "./control/btn_list";
import { QRCodeIcon } from "./visual/qrcode";
import { getErrMsg } from "../common/utils";
import { alertMsg } from "../common/env";
import { updater } from "./state_updater";
import { ICoreState, MsgProps, UIProps } from "./core_state";
import { LoginProps } from "./pane_login";
import { Flexbox } from "./layout/flexbox";
import { Container } from "./layout/container";
import { Rows } from "./layout/rows";
import { NotFoundBanner } from "./visual/banner_notfound";
import { Title } from "./visual/title";
import { shareIDQuery } from "../client/files";
import { loadingCtrl, ctrlOn, ctrlOff } from "../common/controls";
import { CronTable } from "../common/cron";

export interface SharingsProps {
  sharings: Map<string, string>;
  orderBy: string;
  order: boolean;
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

  componentDidMount(): void {
    CronTable().setInterval("listSharings", {
      func: updater().listSharings,
      args: [],
      delay: 5000,
    });
  }

  componentWillUnmount() {
    CronTable().clearInterval("listSharings");
  }

  setLoading = (state: boolean) => {
    updater().setControlOption(loadingCtrl, state ? ctrlOn : ctrlOff);
    this.props.update(updater().updateUI);
  };

  deleteSharing = async (dirPath: string) => {
    this.setLoading(true);

    try {
      const deleteStatus = await updater().deleteSharing(dirPath);
      if (deleteStatus !== "") {
        throw deleteStatus;
      }
      updater().setSharing(false);

      await this.listSharings();
    } catch (e: any) {
      alertMsg(getErrMsg(this.props.msg.pkg, "op.fail", status));
    } finally {
      this.setLoading(false);
    }
  };

  listSharings = async () => {
    this.setLoading(true);
    try {
      const status = await updater().listSharings();
      if (status !== "") {
        alertMsg(getErrMsg(this.props.msg.pkg, "op.fail", status));
      }
      this.props.update(updater().updateFilesInfo);
      this.props.update(updater().updateSharingsInfo);
    } finally {
      this.setLoading(false);
    }
  };

  makeRows = (sharings: Map<string, string>): List<React.ReactNode> => {
    const sharingRows = sharings
      .keySeq()
      .map((dirPath: string): React.ReactNode => {
        const shareID = sharings.get(dirPath);
        const sharingURL = `${
          document.location.href.split("?")[0]
        }?${shareIDQuery}=${shareID}`;

        const row1 = (
          <div>
            <div className="col-l">
              <span className="title-m-wrap dark-font">{dirPath}</span>
            </div>

            <div className="col-r">
              <Flexbox
                children={List([
                  <span className="margin-r-m">
                    <QRCodeIcon value={sharingURL} size={128} pos={false} />
                  </span>,

                  <button
                    onClick={() => {
                      this.deleteSharing(dirPath);
                    }}
                    className="button-default"
                  >
                    {this.props.msg.pkg.get("op.cancel")}
                  </button>,
                ])}
                childrenStyles={List([
                  { flex: "0 0 auto" },
                  { flex: "0 0 auto" },
                ])}
                style={{ justifyContent: "flex-end" }}
              />
            </div>
            <div className="fix"></div>
          </div>
        );

        const elem = (
          <div className="sharing-item" key={dirPath}>
            {row1}
            <div className="desc">{sharingURL}</div>
            <div className="hr"></div>
          </div>
        );

        return elem;
      });

    return sharingRows.toList();
  };

  updateSharings = (sharings: List<Object>) => {
    const newSharingDirs = sharings as List<string>;
    let newSharings = Map<string, string>();
    newSharingDirs.forEach((dirPath) => {
      const shareID = this.props.sharingsInfo.sharings.get(dirPath);
      newSharings = newSharings.set(dirPath, shareID);
    });
    updater().updateSharings(newSharings);
    this.props.update(updater().updateSharingsInfo);
  };

  orderBy = (columnName: string) => {
    const order = !this.props.sharingsInfo.order;
    updater().sortSharings(columnName, order);
    this.props.update(updater().updateSharingsInfo);
  };

  render() {
    const title = (
      <Title
        title={this.props.msg.pkg.get("browser.share.title")}
        iconName="RiShareBoxLine"
        iconColor="black"
      />
    );

    const orderByCallbacks = List([
      () => {
        this.orderBy(this.props.msg.pkg.get("item.path"));
      },
    ]);
    const orderByButtons = (
      <BtnList
        titleIcon="BiSortUp"
        btnNames={List([this.props.msg.pkg.get("item.path")])}
        btnCallbacks={orderByCallbacks}
      />
    );

    const sharingRows = this.makeRows(this.props.sharingsInfo.sharings);
    const view =
      this.props.sharingsInfo.sharings.size > 0 ? (
        <div>
          {orderByButtons}
          <Rows rows={sharingRows} />
        </div>
      ) : (
        <NotFoundBanner title={this.props.msg.pkg.get("share.404.title")} />
      );

    return (
      <div id="sharing-list">
        <Container>
          {title}
          <div className="hr"></div>
          {view}
        </Container>
      </div>
    );
  }
}
