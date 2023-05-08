import "../index.css"
import { Menu } from '../components/atoms/menu.jsx';

export default {
  title: 'Atoms/Menu',
  component: Menu,
  tags: ['autodocs'],
  argTypes: {},
}

export const DefaultMenu = {
  args: {
    items: [
      {label: "Item 1", value: "item1"},
      {label: "Item 2", value: "item2"},
      {label: "Item 3", value: "item3"},
    ],
    value: "item1",
    placeholder: "Select an item"
  }
}