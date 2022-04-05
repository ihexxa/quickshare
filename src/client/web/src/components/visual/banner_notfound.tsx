import * as React from "react";
import { List } from "immutable";

import { RiFileList2Fill } from "@react-icons/all-files/ri/RiFileList2Fill";

import { Flexbox } from "../layout/flexbox";

export interface Props {
  title: string;
}

export const NotFoundBanner = (props: Props) => {
  return (
    <Flexbox
      children={List([
        <RiFileList2Fill size="4rem" className="margin-r-m normal-font" />,
        <span>
          <h3 className="title-l">{props.title}</h3>
        </span>,
      ])}
      childrenStyles={List([
        { flex: "auto", justifyContent: "flex-end" },
        { flex: "auto" },
      ])}
      className="margin-t-l margin-b-l"
    />
  );
};
