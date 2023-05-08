import "../index.css"
import { Button } from '../components/atoms/button.jsx';
import {ButtonGroup} from "../components/atoms/button-groups.jsx";


export default {
  title: 'Molecules/ButtonGroups',
  component: ButtonGroup,
  tags: ['autodocs'],
  argTypes: {},
};


export const SegmentedGroup = {
  args: {
    style: 'basic',
    label: 'Button',
  },
};

export const DefaultGroup = {
  args: {
    style: 'outline',
    label: 'Button',
  },
};
