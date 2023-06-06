import "../../index.css"
import { ArrowsDownUp, CaretDownFill, List, SquaresFour } from "@jengaicons/react";
import { Button } from "../../components/atoms/button";
import { ButtonGroup } from "../../components/atoms/button-groups";
import { Filters } from "../../components/molecule/filters";

export default {
    title: "Molecules/Filters",
    component: Filters,
    tags: ['autodocs'],
    argTypes: {
        filterActions: {
            table: {
                disable: true
            }
        }
    }
}

export const DefaultFilter = {
    args: {
        filterActions: <>
            <ButtonGroup
                items={
                    [
                        {
                            label: "Provider",
                            value: "provider",
                            key: "provider",
                            disclosureComp: CaretDownFill
                        },
                        {
                            label: "Region",
                            value: "region",
                            key: "region",
                            disclosureComp: CaretDownFill
                        },
                        {
                            label: "Status",
                            value: "status",
                            key: "status",
                            disclosureComp: CaretDownFill
                        }
                    ]
                }
            />
            <Button
                label="Sortby"
                IconComp={ArrowsDownUp}
                style={"basic"}
            />
            <ButtonGroup
                selectable
                value={"list"}
                items={
                    [
                        {
                            value: "list",
                            key: "list",
                            icon: List
                        },
                        {
                            value: "grid",
                            key: "grid",
                            icon: SquaresFour
                        }
                    ]
                }
            />
        </>    }
}