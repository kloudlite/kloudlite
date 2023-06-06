import "../index.css"
import { Button } from '../components/atoms/button.jsx';
import { CalendarCheckFill, CaretDownFill } from "@jengaicons/react";


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
    IconComp: CalendarCheckFill,
    DisclosureComp: CaretDownFill,
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