import "../index.css"
import { Chip } from "../components/atoms/chip";


export default {
  title: 'Atoms/Chip',
  component: Chip,
  tags: ['autodocs'],
  argTypes: {},
};


export const SelectedChip = {
  args: {
    selected: true,
  },
};

export const DisabledChip = {
  args: {
    disabled: true,
  },
};

