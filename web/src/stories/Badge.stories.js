import "../index.css"
import {Badge} from "../components/atoms/Badge.jsx";


export default {
  title: 'Atoms/Badge',
  component: Badge,
  tags: ['autodocs'],
  argTypes: {},
};


export const SelectedBadge = {
  args: {
    selected: true,
  },
};

export const DisabledBadge = {
  args: {
    disabled: true,
  },
};

