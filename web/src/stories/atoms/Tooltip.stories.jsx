import "~/lib/app-setup/index.css"
import { Tooltip, TooltipProvider } from "../../components/atoms/tooltip";
import { Button } from "../../components/atoms/button";


export default {
    title: 'Atoms/Tooltip',
    component: Tooltip,
    decorators: [
        (Story) => (
            <TooltipProvider>
                <Story />
            </TooltipProvider>
        ),
    ],
    tags: ['autodocs'],
    argTypes: {},
};

export const InitialAvatar = {
    args: {
        content: "tooltip",
        children: <Button />
    }
}