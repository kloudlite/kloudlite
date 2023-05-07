
import "../index.css"
import {SegmentControl} from "../components/atoms/segment-control";

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
