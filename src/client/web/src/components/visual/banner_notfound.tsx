import * as React from "react";
import { List } from "immutable";

import { RiQuestionnaireFill } from "@react-icons/all-files/ri/RiQuestionnaireFill";

import { Flexbox } from "../layout/flexbox";

export interface Props {
  title: string;
}

export const NotFoundBanner = (props: Props) => {
  return (
    <Flexbox
      children={List([
        <RiQuestionnaireFill size="4rem" className="margin-r-m red0-font" />,
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
