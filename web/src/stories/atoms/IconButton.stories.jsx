import "../../index.css"
import { IconButton } from '../../components/atoms/button.jsx';
import { CalendarCheckFill } from "@jengaicons/react";


export default {
  title: 'Atoms/IconButton',
  component: IconButton,
  tags: ['autodocs'],
  argTypes: {},
};


export const BaseButton = {
  args: {
    variant: 'basic',
    IconComp: CalendarCheckFill,
  },
};