import * as React from "react";

export interface Props {
  name: string;
  value: string;
}

export const Card = (props: Props) => {
  return (
    <div className="padding-m float-l">
      <div className="title-l black0-font">{props.value}</div>
      <div className="desc-m">{props.name}</div>
    </div>
  );
};
