import * as React from "react";
import { List } from "immutable";

import { Flexbox } from "../layout/flexbox";
import { getIconWithProps } from "../visual/icons";

export type BtnListCallBack = () => void;
export interface Props {
  titleIcon?: string;
  btnNames: List<string>;
  btnCallbacks: List<BtnListCallBack>;
}

export const BtnList = (props: Props) => {
  const titleIcon =
    props.titleIcon != null ? (
      getIconWithProps(props.titleIcon, {
        size: "3rem",
        className: "black-font margin-r-m",
      })
    ) : (
      <span></span>
    );

  const btns = props.btnNames.map((btnName: string, i: number) => {
    const cb = props.btnCallbacks.get(i);
    return (
      <button key={`rows-${i}`} className="inline-block margin-r-m button-default" onClick={cb}>
        {btnName}
      </button>
    );
  });

  return (
    <div className="margin-b-l">
      <Flexbox
        children={List([titleIcon, <span>{btns}</span>])}
        childrenStyles={List([{ flex: "0 0 auto" }, { flex: "0 0 auto" }])}
      />
    </div>
  );
};
