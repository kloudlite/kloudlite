import "../index.css"
import { Thumbnail } from "../components/atoms/thumbnail";


export default {
  title: 'Atoms/Thumbnail',
  component: Thumbnail,
  tags: ['autodocs'],
  argTypes: {},
};


export const Normal = {
  args: {
    size: "medium",
    src:"https://images.unsplash.com/photo-1600716051809-e997e11a5d52?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHxzZWFyY2h8NHx8c2FtcGxlfGVufDB8fDB8fA%3D%3D&auto=format&fit=crop&w=800&q=60"
  },
};

export const Rounded = {
  args: {
    size: "medium",
    rounded: true,
    src:"https://images.unsplash.com/photo-1600716051809-e997e11a5d52?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHxzZWFyY2h8NHx8c2FtcGxlfGVufDB8fDB8fA%3D%3D&auto=format&fit=crop&w=800&q=60"
  },
};

