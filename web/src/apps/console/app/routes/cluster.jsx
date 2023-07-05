import { EmptyState } from "~/root/src/components/molecule/empty-state"
import { SubHeader } from "~/root/src/components/organisms/sub-header"

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