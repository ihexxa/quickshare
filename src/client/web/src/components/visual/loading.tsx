import * as React from "react";

import { RiLoader5Fill } from "@react-icons/all-files/ri/RiLoader5Fill";

export interface Props {}

export interface State {}

export const LoadingIcon = (props: Props) => {
  return (
    <div id="loading-container">
      <RiLoader5Fill
        size="5rem"
        className="cyan0-font anm-rotate-s"
        style={{ position: "absolute" }}
      />
      <RiLoader5Fill
        size="5rem"
        className="cyan1-font anm-rotate-m"
        style={{ position: "absolute" }}
      />
      <RiLoader5Fill
        size="5rem"
        className="blue1-font anm-rotate-f"
        style={{ position: "absolute" }}
      />
    </div>
  );
};
