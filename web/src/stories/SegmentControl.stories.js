import { Button } from '../components/atoms/Button.jsx';
import "../index.css"
import {SegmentControl} from "../components/atoms/SegmentControl.jsx";


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
