import "~/lib/app-setup/index.css"
import { ButtonGroup } from "../../components/atoms/button-groups.jsx";
import { BellFill, GearFill } from "@jengaicons/react";


export default {
  title: 'Molecules/ButtonGroups',
  component: ButtonGroup,
  tags: ['autodocs'],
  argTypes: {},
};


export const DefaultButtonGroup = {
  args: {
    size: 'medium',
    onChange: (e) => { console.log(e) },
    value:"item1",
    items: [
      {
        label: "item 1",
        value: "item1",
        key: "item1",
      },
      {
        value: "item2",
        label: "item 2",
        key: "item2",
      },
      {
        label: "item 3",
        value: "item3",
        key: "item3"
      },
      {
        label: "item 4",
        value: "item4",
        key: "item4"
      },
    ]
  },
};

export const DefaultButtonGroupSelectable = {
  args: {
    size: 'medium',
    onChange: (e) => { console.log(e) },
    value:"item1",
    selectable: true,
    items: [
      {
        label: "item 1",
        value: "item1",
        key: "item1",
      },
      {
        value: "item2",
        label: "item 2",
        key: "item2",
      },
      {
        label: "item 3",
        value: "item3",
        key: "item3"
      },
      {
        label: "item 4",
        value: "item4",
        key: "item4"
      },
    ]
  },
};

export const IconButtonGroup = {
  args: {
    size: 'medium',
    onChange: (e) => { console.log(e) },
    value:"item1",
    items: [
      {
        value: "item1",
        key: "item1",
        icon: BellFill
      },
      {
        value: "item2",
        key: "item2",
        icon: GearFill
      }
    ]
  },
};

export const IconButtonGroupSelectable = {
  args: {
    size: 'medium',
    onChange: (e) => { console.log(e) },
    value:"item1",
    selectable: true,
    items: [
      {
        value: "item1",
        key: "item1",
        icon: BellFill
      },
      {
        value: "item2",
        key: "item2",
        icon: GearFill
      }
    ]
  },
};