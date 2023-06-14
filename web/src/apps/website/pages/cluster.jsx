import { EmptyState } from "../../../components/molecule/empty-state.jsx"
import { SubHeader } from "../../../components/organisms/sub-header.jsx"

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