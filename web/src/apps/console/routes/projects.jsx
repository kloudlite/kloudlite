import { useState } from "react"
import { useNavigate } from "@remix-run/react"
import { ArrowsDownUpFill, CaretDownFill, List, PlusFill, SquaresFour } from "@jengaicons/react"
import { SubHeader } from "~/root/src/stories/components/organisms/sub-header.jsx"
import { Button } from "~/root/src/stories/components/atoms/button.jsx"
import { Filters } from "~/root/src/stories/components/molecule/filters.jsx"
import { ButtonGroup } from "~/root/src/stories/components/atoms/button-groups.jsx"
import { EmptyState } from "~/root/src/stories/components/molecule/empty-state.jsx"
import { Tooltip, TooltipProvider } from "~/root/src/stories/components/atoms/tooltip.jsx"


const Projects = ({ }) => {

    const [projects, setProjects] = useState([1])

    const navigate = useNavigate()

    const [projectListMode, setProjectListMode] = useState("list")

    return <>
        <SubHeader title={"Projects"} actions={
            projects.length != 0 && <>
                <Button variant="primary" content="Add new" IconComp={PlusFill} onClick={() => {
                    navigate("../new-project")
                    console.log("called");
                }} />
            </>
        } />
        {projects.length > 0 && <div className="pt-5 flex flex-col gap-10">

            <Filters filterActions={
                <div className="flex flex-row gap-2 items-center justify-center">

                    <TooltipProvider delayDuration={400}>
                        <Tooltip content={"Hello"}>
                            <Button IconComp={ArrowsDownUpFill} variant="basic" content="Sortby" />
                        </Tooltip>
                    </TooltipProvider>
                    <ButtonGroup
                        selectable
                        value={"list"}
                        onChange={(e) => {
                            setProjectListMode(e)
                        }}
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
            <div>
                {/* <ResourceList items={[1, 2, 3, 4, 5]} mode={projectListMode} /> */}
            </div>
        </div>}
        {projects.length == 0 && <div className="pt-5">
            <EmptyState
                heading={"This is where you’ll manage your projects"}
                children={
                    <p>
                        You can create a new project and manage the listed project.
                    </p>
                }
                action={{
                    title: "Create Project",
                    click: () => {
                        navigate("/new-project")
                    }
                }}
            />
        </div>}
    </>
}

export default Projects