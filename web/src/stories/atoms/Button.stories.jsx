import "~/lib/app-setup/index.css"
import { Button } from '../../components/atoms/button.jsx';
import { CalendarCheckFill, CaretDownFill } from "@jengaicons/react";


export default {
  title: 'Atoms/Button',
  component: Button,
  tags: ['autodocs'],
  argTypes: {},
};


export const BaseButton = {
  args: {
    variant: 'basic',
    label: 'Button',
    IconComp: CalendarCheckFill,
    DisclosureComp: CaretDownFill,
  },
};

export const OutlineButton = {
  args: {
    variant: 'outline',
    label: 'Button',
  },
};

export const PlainButton = {
  args: {
    variant: 'plain',
    label: 'Button',
  },
};