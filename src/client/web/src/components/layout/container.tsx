import * as React from "react";

export interface Props {
  children: React.ReactNode | undefined;
}

export const Container = (props: Props) => {
  return (
    <div className="container">
      <div className="container-padding">{props.children}</div>
    </div>
  );
};
