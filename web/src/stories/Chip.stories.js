import "../index.css"
import { Chip } from "../components/atoms/chip";
import { CalendarCheckFill, CaretDownFill } from "@jengaicons/react";


export default {
  title: 'Atoms/Chip',
  component: Chip,
  tags: ['autodocs'],
  argTypes: {},
};


export const SelectedChip = {
  args: {
    selected: true,
    LeftIconComp: CalendarCheckFill,
  },
};

export const DisabledChip = {
  args: {
    disabled: true,
  },
};

