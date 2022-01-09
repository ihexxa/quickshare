import * as React from "react";
import QRCode from "react-qr-code";

import { RiQrCodeFill } from "@react-icons/all-files/ri/RiQrCodeFill";

export interface Props {
  value: string;
  size: number;
  pos: boolean; // true=left, right=false
  className?: string;
}

export interface State {
  show: boolean;
}

export class QRCodeIcon extends React.Component<Props, State, {}> {
  constructor(p: Props) {
    super(p);
    this.state = {
      show: false,
    };
  }

  toggle = () => {
    this.setState({ show: !this.state.show });
  };

  show = () => {
    this.setState({ show: true });
  };
  hide = () => {
    this.setState({ show: false });
  };

  render() {
    const widthInRem = `${Math.floor(this.props.size / 10)}rem`;
    const posStyle = this.props.pos
      ? { left: `0` }
      : { left: `-${widthInRem}` };
    const qrcode = this.state.show ? (
      <div className="qrcode-child" style={{ ...posStyle }}>
        <div className="qrcode">
          <QRCode value={this.props.value} size={this.props.size} />
        </div>
      </div>
    ) : null;

    return (
      <div className={`qrcode-container ${this.props.className}`}>
        <RiQrCodeFill
          className="qrcode-icon"
          onMouseEnter={this.show}
          onMouseLeave={this.hide}
          onClick={this.toggle}
          size={"2rem"}
        />
        <div className="qrcode-child-container">{qrcode}</div>
      </div>
    );
  }
}
