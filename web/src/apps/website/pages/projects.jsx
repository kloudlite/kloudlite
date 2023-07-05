import { useState } from "react"
import { EmptyState } from "~/root/src/stories/components/molecule/empty-state.jsx"
import { SubHeader } from "~/root/src/stories/components/organisms/sub-header.jsx"
import { Button } from "~/root/src/stories/components/atoms/button.jsx"
import { ArrowsDownUpFill, CaretDownFill, List, PlusFill, SquaresFour } from "@jengaicons/react"
import { Filters } from "~/root/src/stories/components/molecule/filters.jsx"
import { ButtonGroup } from "~/root/src/stories/components/atoms/button-groups.jsx"

const Projects = ({ }) => {

    const [projects, setProjects] = useState([0])

    return <>
        <SubHeader title={"Projects"} actions={
            projects.length != 0 && <>
                <Button style="primary" label="Add new" IconComp={PlusFill} />
            </>
        } />

        <Filters filterActions={
            <div className="flex flex-row gap-2 items-center justify-center">
                <ButtonGroup items={[
                    {
                        label: "Status",
                        key: "status",
                        value: "status",
                        disclosureComp: CaretDownFill
                    },
                    {
                        label: "Cluster",
                        key: "cluster",
                        value: "cluster",
                        disclosureComp: CaretDownFill
                    }
                ]} />
                <Button IconComp={ArrowsDownUpFill} style="basic" label="Sortby" />
                <ButtonGroup
                    selectable
                    value={"list"}
                    items={[
                        {
                            key: "list",
                            value: "list",
                            icon: List
                        },
                        {
                            key: "grid",
                            value: "grid",
                            icon: SquaresFour
                        }
                    ]} />
            </div>
        } />
        {projects.length == 0 && <EmptyState
            heading={"This is where you’ll manage your projects"}
            children={
                <p>
                    You can create a new project and manage the listed project.
                </p>
            }
            action={{
                title: "Create Project"
            }}
        />}
    </>
}

export default Projects