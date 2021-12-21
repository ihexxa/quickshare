import * as React from "react";
import { Map } from "immutable";

import { IconType, IconBaseProps } from "@react-icons/all-files";
import { RiFolder2Fill } from "@react-icons/all-files/ri/RiFolder2Fill";
import { RiShareBoxLine } from "@react-icons/all-files/ri/RiShareBoxLine";
import { RiUploadCloudFill } from "@react-icons/all-files/ri/RiUploadCloudFill";
import { RiSettings3Fill } from "@react-icons/all-files/ri/RiSettings3Fill";
import { RiWindowFill } from "@react-icons/all-files/ri/RiWindowFill";
import { RiCheckboxBlankFill } from "@react-icons/all-files/ri/RiCheckboxBlankFill";
import { RiCheckboxFill } from "@react-icons/all-files/ri/RiCheckboxFill";
import { RiMenuFill } from "@react-icons/all-files/ri/RiMenuFill";
import { RiInformationFill } from "@react-icons/all-files/ri/RiInformationFill";
import { RiDeleteBin2Fill } from "@react-icons/all-files/ri/RiDeleteBin2Fill";
import { RiArchiveDrawerFill } from "@react-icons/all-files/ri/RiArchiveDrawerFill";
import { RiFileList2Fill } from "@react-icons/all-files/ri/RiFileList2Fill";
import { RiArrowUpDownFill } from "@react-icons/all-files/ri/RiArrowUpDownFill";
import { BiTable } from "@react-icons/all-files/bi/BiTable";
import { BiListUl } from "@react-icons/all-files/bi/BiListUl";

import { colorClass } from "./colors";

export interface IconProps {
  name: string;
  size: string;
  color: string;
}

const icons = Map<string, IconType>({
  RiFileList2Fill: RiFileList2Fill,
  RiFolder2Fill: RiFolder2Fill,
  RiShareBoxLine: RiShareBoxLine,
  RiUploadCloudFill: RiUploadCloudFill,
  RiSettings3Fill: RiSettings3Fill,
  RiWindowFill: RiWindowFill,
  RiCheckboxBlankFill: RiCheckboxBlankFill,
  RiCheckboxFill: RiCheckboxFill,
  RiMenuFill: RiMenuFill,
  RiInformationFill: RiInformationFill,
  RiDeleteBin2Fill: RiDeleteBin2Fill,
  RiArchiveDrawerFill: RiArchiveDrawerFill,
  RiArrowUpDownFill: RiArrowUpDownFill,
  BiTable: BiTable,
  BiListUl: BiListUl,
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
