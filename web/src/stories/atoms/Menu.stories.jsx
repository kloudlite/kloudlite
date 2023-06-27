import "../../index.css"
import { Menu } from '../../components/atoms/menu.jsx';
import { Button } from "../../components/atoms/button";

export default {
  title: 'Atoms/Menu',
  component: Menu,
  tags: ['autodocs'],
  argTypes: {},
}

// export const DefaultMenu = {
//   args: {
//     items: [
//       { value: "Item-1", Element: <div className="w-32">Hi</div> },
//       { value: "Item-2", Element: <div className="w-32">Hi2</div> },
//       { value: "Item-3", Element: <div className="w-32">Hi3</div> },
//     ],
//     children: [
//       <Button key={1} label="Profile" />,
//       <Button key={2} label="Profile" />
//     ]
//   }
// }