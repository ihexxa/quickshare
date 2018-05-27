import React from "react";

const buttonClassName = "btn";

const styleContainer = {
  display: "inline-block",
  height: "2.5rem"
};

const styleIcon = {
  lineHeight: "2.5rem",
  height: "2.5rem",
  margin: "0 -0.25rem 0 0.5rem"
};

const styleBase = {
  background: "transparent",
  lineHeight: "2.5rem",
  fontSize: "0.875rem",
  border: "none",
  outline: "none",
  padding: "0 0.75rem",
  textAlign: "center"
};

const styleDefault = {
  ...styleBase
};

const styleStr = `
  .${buttonClassName}:hover {
    opacity: 0.7;
    transition: opacity 0.25s;
  }

  .${buttonClassName}:active {
    opacity: 0.7;
    transition: opacity 0.25s;
  }

  .${buttonClassName}:disabled {
    opacity: 0.2;
    transition: opacity 0.25s;
  }

  button::-moz-focus-inner {
    border: 0;
  }
`;

export class Button extends React.PureComponent {
  constructor(props) {
    super(props);
    this.styleDefault = { ...styleDefault, ...this.props.styleDefault };
    this.styleStr = this.props.styleStr ? this.props.styleStr : styleStr;
  }

  onClick = e => {
    if (this.props.onClick && this.props.isEnabled) {
      this.props.onClick(e);
    }
  };

  render() {
    const style = this.props.isEnabled ? this.styleDefault : this.styleDisabled;
    const icon =
      this.props.icon != null ? (
        <span style={{ ...styleIcon, ...this.props.styleIcon }}>
          {this.props.icon}
        </span>
      ) : (
        <span />
      );

    return (
      <div
        style={{ ...styleContainer, ...this.props.styleContainer }}
        className={`${buttonClassName} ${this.props.className}`}
        onClick={this.onClick}
      >
        {icon}
        <button style={style}>
          <span style={this.props.styleLabel}>{this.props.label}</span>
          <style>{this.styleStr}</style>
        </button>
      </div>
    );
  }
}

Button.defaultProps = {
  className: "btn",
  isEnabled: true,
  icon: null,
  onClick: () => true,
  styleContainer: {},
  styleDefault: {},
  styleDisabled: {},
  styleLabel: {},
  styleIcon: {},
  styleStr: undefined
};
