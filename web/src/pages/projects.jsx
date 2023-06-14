import { useState } from "react"
import { EmptyState } from "../components/molecule/empty-state"
import { SubHeader } from "../components/organisms/sub-header"
import { Button } from "../components/atoms/button"
import { ArrowsDownUpFill, CaretDownFill, List, PlusFill, SquaresFour } from "@jengaicons/react"
import { Filters } from "../components/molecule/filters"
import { ButtonGroup } from "../components/atoms/button-groups"
import { useNavigate } from "react-router-dom"

const Projects = ({ }) => {

    const [projects, setProjects] = useState([])
    const navigate = useNavigate()
    return <>
        {projects.length > 0 && <>
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
        </>}
        {projects.length == 0 && <EmptyState
            heading={"This is where you’ll manage your projects"}
            children={
                <p>
                    You can create a new project and manage the listed project.
                </p>
            }
            action={{
                title: "Create Project",
                click: () => {
                    navigate("/main/newproject")
                }
            }}
        />}
    </>
}

export default Projects