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
        size: "1.8rem",
        className: "major-font mr-4",
      })
    ) : (
      <span></span>
    );

  const btns = props.btnNames.map((btnName: string, i: number) => {
    const cb = props.btnCallbacks.get(i);
    const isLast = i === props.btnNames.size - 1;
    return (
      <button
        key={`rows-${i}`}
        className={`inline-block ${isLast ? "" : "mr-8"}`}
        onClick={cb}
      >
        {btnName}
      </button>
    );
  });

  return (
    <div>
      <Flexbox
        children={List([titleIcon, <span>{btns}</span>])}
        childrenStyles={List([{ flex: "0 0 auto" }, { flex: "0 0 auto" }])}
      />
    </div>
  );
};
