import React from "react";

const styleContainer = {
  backgroundColor: "#ccc",
  display: "inline-block",
  height: "2.5rem"
};

const styleIcon = {
  lineHeight: "2.5rem",
  height: "2.5rem",
  margin: "0 0.25rem 0 0.5rem"
};

const styleInputBase = {
  backgroundColor: "transparent",
  border: "none",
  display: "inline-block",
  fontSize: "0.875rem",
  height: "2.5rem",
  lineHeight: "2.5rem",
  outline: "none",
  overflowY: "hidden",
  padding: "0 0.75rem",
  verticalAlign: "middle"
};

const styleDefault = {
  ...styleInputBase,
  color: "#333"
};

const styleInvalid = {
  ...styleInputBase,
  color: "#e74c3c"
};

const inputClassName = "qs-input";
const styleStr = `
.${inputClassName}:hover {
  // box-shadow: 0px 0px -5px rgba(0, 0, 0, 1);
  opacity: 0.7;
  transition: opacity 0.25s;
}

.${inputClassName}:active {
  // box-shadow: 0px 0px -5px rgba(0, 0, 0, 1);
  opacity: 0.7;
  transition: opacity 0.25s;
}

.${inputClassName}:disabled {
  color: #ccc;
}
`;

export class Input extends React.PureComponent {
  constructor(props) {
    super(props);
    this.state = { isValid: true };
    this.inputRef = undefined;
  }

  onChange = e => {
    this.props.onChange(e.target.value);
    this.props.onChangeEvent(e);
    this.props.onChangeTarget(e.target);
    this.setState({ isValid: this.props.validate(e.target.value) });
  };

  getRef = input => {
    this.inputRef = input;
    this.props.inputRef(this.inputRef);
  };

  render() {
    const style = this.state.isValid ? styleDefault : styleInvalid;
    const icon =
      this.props.icon != null ? (
        <span style={styleIcon}>{this.props.icon}</span>
      ) : (
        <span />
      );

    return (
      <div style={{ ...styleContainer, ...this.props.styleContainer }}>
        {icon}
        <input
          style={{
            ...styleDefault,
            ...this.props.style,
            width: this.props.width
          }}
          className={`${inputClassName} ${this.props.className}`}
          disabled={this.props.disabled}
          readOnly={this.props.readOnly}
          maxLength={this.props.maxLength}
          placeholder={this.props.placeholder}
          type={this.props.type}
          onChange={this.onChange}
          value={this.props.value}
          ref={this.getRef}
        />
        <style>{styleStr}</style>
      </div>
    );
  }
}

Input.defaultProps = {
  className: "input",
  maxLength: "32",
  placeholder: "placeholder",
  readOnly: false,
  style: {},
  styleContainer: {},
  styleInvalid: {},
  type: "text",
  disabled: false,
  width: "auto",
  value: "",
  icon: null,
  onChange: () => true,
  onChangeEvent: () => true,
  onChangeTarget: () => true,
  validate: () => true,
  inputRef: () => true
};
