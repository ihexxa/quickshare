import * as React from "react";

export interface Props {
  isHorizontal: boolean;
  elements: Array<JSX.Element>;
}

export interface State {}
export class Layouter extends React.Component<Props, State, {}> {
  constructor(p: Props) {
    super(p);
  }

  horizontalLayout = (children: Array<JSX.Element>): Array<JSX.Element> => {
    return children.map((child: JSX.Element, idx: number) => {
      // if (idx === 0) {
      //   return <span key={`layout=${idx}`}>{child}</span>;
      // }
      return (
        <span key={`layout=${idx}`}>
          {child}
          <span className="margin-s"></span>
        </span>
      );
    });
  };

  verticalLayout = (children: Array<JSX.Element>): Array<JSX.Element> => {
    return this.horizontalLayout(children);
  };

  render() {
    const elements = this.props.isHorizontal
      ? this.horizontalLayout(this.props.elements)
      : this.verticalLayout(this.props.elements);
    return <div className="layouter">{elements}</div>;
  }
}
