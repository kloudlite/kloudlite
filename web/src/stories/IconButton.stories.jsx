import "../index.css"
import { IconButton } from '../components/atoms/button.jsx';
import { CalendarCheck, CalendarCheckFill, CaretDownFill } from "@jengaicons/react";


export default {
  title: 'Atoms/IconButton',
  component: IconButton,
  tags: ['autodocs'],
  argTypes: {},
};


export const BaseButton = {
  args: {
    style: 'basic',
    IconComp: CalendarCheckFill,
  },
};