import { EmptyState } from "~/root/src/stories/components/molecule/empty-state.jsx"
import { SubHeader } from "~/root/src/stories/components/organisms/sub-header.jsx"

export default () => {
    return <>
        <SubHeader title={"Cluster"} />
        <EmptyState
            heading={"This is where you’ll manage your cluster "}
            children={
                <p>
                    You can create a new cluster and manage the listed cluster.
                </p>
            }
            action={{
                title: "Create Cluster"
            }}
        />
    </>
}