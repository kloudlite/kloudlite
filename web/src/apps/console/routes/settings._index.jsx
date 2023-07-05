import { redirect } from "@remix-run/node";

export const loader = async () => {
    return redirect("/settings/general");

};