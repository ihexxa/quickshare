import React from "react";
import { Button } from "../control/button";
import { Input } from "../control/input";

export const classDelBtn = "file-box-pane-btn-del";
export const classDelYes = "del-no";
export const classDelNo = "del-yes";

let styleDetailPane = {
  color: "#666",
  backgroundColor: "#fff",
  position: "absolute",
  marginBottom: "5rem",
  zIndex: "10"
};

const styleDetailContainer = {
  padding: "1em",
  borderBottom: "solid 1rem #ccc"
};

const styleDetailHeader = {
  color: "#999",
  fontSize: "0.75rem",
  fontWeight: "bold",
  margin: "1.5rem 0 0.5rem 0",
  padding: 0,
  textTransform: "uppercase"
};

const styleDesc = {
  overflow: "hidden",
  whiteSpace: "nowrap",
  textOverflow: "ellipsis",
  lineHeight: "1.5rem",
  fontSize: "0.875rem"
};

export class FileBoxDetail extends React.PureComponent {
  constructor(props) {
    super(props);
    this.state = {
      downLimit: this.props.downLimit,
      showDelComfirm: false
    };

    styleDetailPane = {
      ...styleDetailPane,
      width: this.props.width
    };
  }

  onResetLink = () => {
    return this.props.onPublishId(this.props.id).then(resettedId => {
      if (resettedId == null) {
        this.props.onError("Resetting link failed");
      } else {
        this.props.onOk("Link is reset");
        this.props.onRefresh();
      }
    });
  };

  onShadowLink = () => {
    return this.props.onShadowId(this.props.id).then(shadowId => {
      if (shadowId == null) {
        this.props.onError("Shadowing link failed");
      } else {
        this.props.onOk("Link is shadowed");
        this.props.onRefresh();
      }
    });
  };

  onSetDownLimit = newValue => {
    this.setState({ downLimit: newValue });
  };

  onComfirmDel = () => {
    this.setState({ showDelComfirm: true });
  };

  onCancelDel = () => {
    this.setState({ showDelComfirm: false });
  };

  onUpdateDownLimit = () => {
    return this.props
      .onSetDownLimit(this.props.id, this.state.downLimit)
      .then(ok => {
        if (ok) {
          this.props.onOk("Download limit updated");
          this.props.onRefresh();
        } else {
          this.props.onError("Setting download limit failed");
        }
      });
  };

  onDelete = () => {
    return this.props.onDel(this.props.id).then(ok => {
      if (ok) {
        this.props.onOk("File deleted");
        this.props.onRefresh();
      } else {
        this.props.onError("Fail to delete file");
      }
    });
  };

  render() {
    const delComfirmButtons = (
      <div>
        <Button
          className={classDelYes}
          label={"DON'T delete"}
          styleContainer={{ backgroundColor: "#2c3e50", marginTop: "0.25rem" }}
          styleDefault={{ color: "#fff" }}
          onClick={this.onCancelDel}
        />
        <Button
          className={classDelNo}
          label={"DELETE it"}
          styleContainer={{ backgroundColor: "#e74c3c", marginTop: "0.25rem" }}
          styleDefault={{ color: "#fff" }}
          onClick={this.onDelete}
        />
      </div>
    );

    return (
      <div style={styleDetailPane} className={this.props.className}>
        <div style={styleDetailContainer}>
          <div>
            <h4 style={styleDetailHeader}>File Information</h4>
            <div>
              <div style={styleDesc}>
                <b>Name</b> {this.props.name}
              </div>
              <div style={styleDesc}>
                <b>Size</b> {this.props.size}
              </div>
              <div style={styleDesc}>
                <b>Time</b> {this.props.modTime}
              </div>
            </div>
          </div>
          <div>
            <h4 style={styleDetailHeader}>Download Link</h4>
            <Input
              type="text"
              value={`${window.location.protocol}//${window.location.host}${
                this.props.href
              }`}
              style={{ marginBottom: "0.5rem" }}
            />
            {/* <Button label={"Copy"} onClick={this.onCopyLink} /> */}
            <br />
            <Button
              label={"Reset"}
              onClick={this.onResetLink}
              styleContainer={{
                backgroundColor: "#ccc",
                marginRight: "0.5rem",
                marginTop: "0.25rem"
              }}
            />
            <Button
              label={"Regenerate"}
              onClick={this.onShadowLink}
              styleContainer={{ backgroundColor: "#ccc", marginTop: "0.25rem" }}
            />
          </div>
          <div>
            <h4 style={styleDetailHeader}>
              Download Limit (-1 means unlimited)
            </h4>
            <Input
              type="text"
              value={this.state.downLimit}
              onChange={this.onSetDownLimit}
              style={{ marginBottom: "0.5rem" }}
            />
            <br />
            <Button
              label={"Update"}
              styleContainer={{ backgroundColor: "#ccc", marginTop: "0.25rem" }}
              onClick={this.onUpdateDownLimit}
            />
          </div>
          <div>
            <h4 style={styleDetailHeader}>Delete</h4>
            {this.state.showDelComfirm ? (
              delComfirmButtons
            ) : (
              <Button
                className={classDelBtn}
                label={"Delete"}
                styleContainer={{
                  backgroundColor: "#e74c3c",
                  marginTop: "0.25rem"
                }}
                styleDefault={{ color: "#fff" }}
                onClick={this.onComfirmDel}
              />
            )}
          </div>
        </div>
      </div>
    );
  }
}

FileBoxDetail.defaultProps = {
  id: "n/a",
  name: "n/a",
  size: "n/a",
  modTime: 0,
  href: "n/a",
  downLimit: -3,
  width: -1,
  className: "",
  onRefresh: () => console.error("undefined"),
  onError: () => console.error("undefined"),
  onOk: () => console.error("undefined"),
  onDel: () => console.error("undefined"),
  onPublishId: () => console.error("undefined"),
  onShadowId: () => console.error("undefined"),
  onSetDownLimit: () => console.error("undefined")
};
