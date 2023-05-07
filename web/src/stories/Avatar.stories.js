import "../index.css"
import {Avatar} from "../components/atoms/Avatar.jsx";


export default {
  title: 'Atoms/Avatar',
  component: Avatar,
  tags: ['autodocs'],
  argTypes: {},
};

export const InitialAvatar = {
  args:{
    size: "medium"
  }
}

export const BasicAvatar = {
  args: {
    label: "Karthik Th",
    size: "medium"
  },
};

