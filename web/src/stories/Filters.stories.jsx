import { Star } from "@jengaicons/react";
import { Button, IconButton } from "../components/atoms/button";
import { ButtonGroup } from "../components/atoms/button-groups";
import { Filters } from "../components/molecule/filters";

export default {
    title: "Molecules/Filters",
    component: Filters,
    tags: ['autodocs'],
    argTypes: {}
}

export const DefaultFilter = {
    args:{
        filterActions:[
            <Button label="hello" style={'outline'}/>,
            <ButtonGroup style={'outline'}/>,
            <IconButton style={"outline"} IconComp={Star}/>
        ]
    }
}