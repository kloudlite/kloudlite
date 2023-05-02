import { Button } from './Button';
import "../index.css"


export default {
  title: 'Atoms/Button',
  component: Button,
  tags: ['autodocs'],
  argTypes: {},
};


export const BaseButton = {
  args: {
    style: 'basic',
    label: 'Button',
  },
};

export const OutlineButton = {
  args: {
    style: 'outline',
    label: 'Button',
  },
};

export const PlainButton = {
  args: {
    style: 'plain',
    label: 'Button',
  },
};
