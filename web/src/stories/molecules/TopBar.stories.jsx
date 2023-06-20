import { createRemixStub } from "@remix-run/testing/dist/create-remix-stub";
import { TopBar } from "../../components/organisms/top-bar";

export default {
  title: "Organisms/Top bar",
  component: TopBar,
  decorators: [
    (Story) => {
      const RemixStub = createRemixStub([
        {
          path: '/',
          element: <Story />,
        },
      ]);

      return <RemixStub />;
    },
  ],
  tags: ['autodocs'],
}


export const DefaultTopBar = {
  args: {

  }
}