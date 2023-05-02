import { Button } from './Button';
import "../index.css"
import {SegmentControl} from "./SegmentControl.jsx";


export default {
  title: 'Atoms/SegmentControl',
  component: SegmentControl,
  tags: ['autodocs'],
  argTypes: {},
};


export const PrimarySegmentControl = {
  args: {
    style: 'primary'
  },
};

export const SecondarySegmentControl = {
  args: {
    style: 'secondary',
  },
};

export const CriticalSegmentControl = {
  args: {
    style: 'critical',
  },
};
