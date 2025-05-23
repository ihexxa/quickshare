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
import { RiMore2Fill } from "@react-icons/all-files/ri/RiMore2Fill";
import { RiCheckboxBlankLine } from "@react-icons/all-files/ri/RiCheckboxBlankLine";
import { BiSortUp } from "@react-icons/all-files/bi/BiSortUp";
import { RiListSettingsFill } from "@react-icons/all-files/ri/RiListSettingsFill";
import { RiHardDriveFill } from "@react-icons/all-files/ri/RiHardDriveFill";
import { RiGridFill } from "@react-icons/all-files/ri/RiGridFill";
import { RiFolderUploadFill } from "@react-icons/all-files/ri/RiFolderUploadFill";
import { RiFileSearchFill } from "@react-icons/all-files/ri/RiFileSearchFill";

import { colorClass } from "./colors";

export interface IconProps {
  name: string;
  size: string;
  color: string;
}

const icons = Map<string, IconType>({
  RiFileList2Fill,
  RiFolder2Fill,
  RiShareBoxLine,
  RiUploadCloudFill,
  RiSettings3Fill,
  RiWindowFill,
  RiCheckboxBlankFill,
  RiCheckboxFill,
  RiMenuFill,
  RiInformationFill,
  RiDeleteBin2Fill,
  RiArchiveDrawerFill,
  RiArrowUpDownFill,
  BiTable,
  BiListUl,
  RiMore2Fill,
  RiCheckboxBlankLine,
  BiSortUp,
  RiListSettingsFill,
  RiHardDriveFill,
  RiGridFill,
  RiFolderUploadFill,
  RiFileSearchFill,
});

export function getIconWithProps(
  name: string,
  props: IconBaseProps
): React.ReactElement<IconBaseProps> | null {
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
): React.ReactElement<IconBaseProps> | null {
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

export function iconSize(size: string): string {
  switch (size) {
    case "s":
      return "2rem";
    case "m":
      return "2.4rem";
    case "l":
      return "3.2rem";
    default:
      throw Error(`icons size(${size}) not found`);
  }
}
