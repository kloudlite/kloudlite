import "../../index.css"
import { Chip, ChipTypes } from "../../components/atoms/chip";
import { CalendarCheckFill, MagicWandFill } from "@jengaicons/react";


export default {
  title: 'Atoms/Chip',
  component: Chip,
  tags: ['autodocs'],
  argTypes: {},
};


export const ChipBasic = {
  args: {
    label: "label",
    prefix: MagicWandFill
  },
};

export const ChipClickable = {
  args: {
    prefix: "title:",
    label: "label",
    type: ChipTypes.CLICKABLE
  },
};

export const ChipRemovable = {
  args: {
    prefix: "Title:",
    label: "label",
    type: ChipTypes.REMOVABLE
  },
};

export const DisabledChip = {
  args: {
    disabled: true
  },
};

