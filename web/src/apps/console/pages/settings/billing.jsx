import { Avatar } from "../../../../components/atoms/avatar"
import { Button } from "../../../../components/atoms/button"
import { TextInput } from "../../../../components/atoms/input"

const BillingSettings = ({ }) => {
    return <div className="flex-1 flex flex-col gap-10">
        <div className="border border-border-default rounded-md p-5 flex flex-col gap-5">
            <span className="text-text-default headingMd">Configure</span>
            <div className="flex flex-row gap-x-5">
                <TextInput label={"Company name"} placeholder={"Company name"} className={"flex-1"} />
                <TextInput label={"Invoice email recipient"} placeholder={"Invoice email recipient"} className={"flex-1"} />
            </div>
        </div>
        <div className="border border-border-danger rounded-md p-5 flex flex-col gap-5">
            <span className="text-text-default headingMd">Delete Account</span>
            <p>
                Permanently remove your personal account and all of its contents from the Kloudlite platform. This action is not reversible, so please continue with caution.
            </p>
            <Button style="critical" label="Delete" />
        </div>
    </div>
}

export default BillingSettings