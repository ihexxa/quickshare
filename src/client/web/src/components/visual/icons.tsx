import * as React from "react";
import { Map } from "immutable";

import { IconType, IconBaseProps } from "@react-icons/all-files";
import { RiFolder2Fill } from "@react-icons/all-files/ri/RiFolder2Fill";
import { RiShareBoxLine } from "@react-icons/all-files/ri/RiShareBoxLine";
import { RiUploadCloudFill } from "@react-icons/all-files/ri/RiUploadCloudFill";

import { colorClass } from "./colors";

export interface IconProps {
  name: string;
  size: string;
  color: string;
}

const icons = Map<string, IconType>({
  RiFolder2Fill: RiFolder2Fill,
  RiShareBoxLine: RiShareBoxLine,
  RiUploadCloudFill: RiUploadCloudFill,
});

export function getIconWithProps(
  name: string,
  props: IconBaseProps
): JSX.Element | null {
  const icon = icons.get(name);
  if (icon == null) {
    throw Error(`icon "${name}" is not found`);
  }

  return React.createElement(icon, { ...props }, null);
}

export function getIcon(
  name: string,
  size: string,
  color: string
): JSX.Element | null {
  const icon = icons.get(name);
  if (icon == null) {
    throw Error(`icon "${name}" is not found`);
  }

  return React.createElement(
    icon,
    { size, className: `${colorClass(color)}-font` },
    null
  );
}
