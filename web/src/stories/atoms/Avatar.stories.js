import "~/lib/app-setup/index.css"
import {Avatar} from "../../components/atoms/avatar";


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

