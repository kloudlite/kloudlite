import { EmptyState } from "../../../components/molecule/empty-state"
import { SubHeader } from "../../../components/organisms/sub-header"

const Cluster = ({ }) => {
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

export default Cluster