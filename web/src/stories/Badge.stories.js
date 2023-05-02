import { Button } from './Button';
import "../index.css"
import {SegmentControl} from "./SegmentControl.jsx";
import {Badge} from "./Badge.jsx";


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

