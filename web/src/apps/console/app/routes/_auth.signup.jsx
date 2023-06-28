import classNames from "classnames";
import { BrandLogo } from "../../../../components/branding/brand-logo";
import { Button } from "../../../../components/atoms/button";
import { ArrowLeft, Envelope, EnvelopeFill, GithubLogoFill, GitlabLogoFill, GoogleLogo } from "@jengaicons/react";
import { useSearchParams } from "react-router-dom";
import { PasswordInput, TextInput } from "../../../../components/atoms/input";

export default function AuthLogin({ }) {
    const [searchParams, setSearchParams] = useSearchParams()
    return <div className={classNames("flex flex-col items-center justify-center min-h-full")}>
        <div className={classNames("flex flex-1 flex-col items-center self-stretch justify-center px-5 py-16 border-b border-border-default")}>
            <div className="flex flex-col items-stretch justify-center gap-8 md:w-[400px]">
                <BrandLogo darkBg={false} size={60} />
                <div className="flex flex-col items-stretch gap-8 border-b pb-8 border-border-default">
                    <div className="flex flex-col gap-2 items-center md:px-12">
                        <div className={classNames("text-text-strong heading3xl text-center")}>Signup to Kloudlite</div>
                        <div className="text-text-soft bodySm text-center">To access your DevOps console, Please provide your login credentials.</div>
                    </div>
                    {searchParams.get('mode') == "email"
                        ?
                        <div className="flex flex-col items-stretch gap-5">
                            <TextInput label={"Name"} placeholder={"Full name"} />
                            <div className="flex flex-col gap-5 items-stretch md:flex-row">
                                <TextInput label={"Company Name"} className={"flex-1"} />
                                {/* <NumberInput label={"Company Size"} className={"flex-1"} min={1} /> */}
                            </div>
                            <TextInput label={"Email"} placeholder={"ex: john@company.com"} />
                            <PasswordInput label={"Password"} placeholder={"XXXXXX"} />
                            <Button size={"large"} variant={"primary"} label="Continue with Email" IconComp={EnvelopeFill} block />
                        </div>
                        :
                        <div className="flex flex-col items-stretch gap-5">
                            <Button size={"large"} variant={"basic"} label="Continue with GitHub" IconComp={GithubLogoFill} href={"https://google.com"} block />
                            <Button size={"large"} variant={"primary"} label="Continue with GitLab" IconComp={GitlabLogoFill} block />
                            <Button size={"large"} variant={"secondary"} label="Continue with Google" IconComp={GoogleLogo} block />
                        </div>}
                </div>
                {searchParams.get('mode') == "email"
                    ?
                    <Button size={"large"} variant={"outline"} label="Other Signup options" IconComp={ArrowLeft} href={"/signup"} block />
                    :
                    <Button size={"large"} variant={"outline"} label="Signup with Email" IconComp={Envelope} href={"/signup/?mode=email"} block />}

                <div className="bodyMd text-text-soft text-center">
                    By signing up, you agree to the Terms of Service and Privacy Policy.
                </div>

            </div>
        </div>
        <div className="py-8  px-5 flex flex-row items-center justify-center self-stretch">
            <div className="bodyMd text-text-default">Already have an account?</div>
            <Button label={"Login"} variant={"primary-plain"} size="medium" href={"/login"} />
        </div>
    </div>
}