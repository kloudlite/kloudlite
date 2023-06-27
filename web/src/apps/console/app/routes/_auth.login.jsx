import classNames from "classnames";
import { BrandLogo } from "../../../../components/branding/brand-logo";
import { Button } from "../../../../components/atoms/button";
import { ArrowLeft, Envelope, EnvelopeFill, GithubLogoFill, GitlabLogoFill, GoogleLogo } from "@jengaicons/react";
import { useSearchParams } from "react-router-dom";
import { PasswordInput, TextInput } from "../../../../components/atoms/input";

const CustomGoogleIcon = (props) => {
    return <GoogleLogo {...props} weight={4}></GoogleLogo>
}

export default function AuthLogin({ }) {
    const [searchParams, setSearchParams] = useSearchParams()
    return <div className={classNames("flex flex-col items-center justify-center h-full")}>
        <div className={classNames("flex flex-1 flex-col items-center self-stretch justify-center px-5 pb-8 border-b border-border-default md:py-37.5")}>
            <div className="flex flex-col items-stretch justify-center gap-8 md:w-[400px]">
                <BrandLogo darkBg={false} size={60} />
                <div className="flex flex-col items-stretch gap-8 border-b pb-8 border-border-default">
                    <div className="flex flex-col gap-2 items-center md:px-12">
                        <div className={classNames("text-text-strong heading3xl text-center")}>Login to Kloudlite</div>
                        <div className="text-text-soft bodySm text-center">To access your DevOps console, Please provide your login credentials.</div>
                    </div>
                    {searchParams.get('mode') == "email"
                        ?
                        <div className="flex flex-col items-stretch gap-5">
                            <TextInput label={"Email"} placeholder={"zuko@example.com"} />
                            <PasswordInput label={"Password"} placeholder={"XXXXXX"} extra={<Button size={"medium"} variant={"primary-plain"} label={"Forgot password"} href={"/forgotpassword"} />} />
                            <Button size={"large"} variant={"primary"} label="Continue with Email" IconComp={EnvelopeFill} />
                        </div>
                        :
                        <div className="flex flex-col items-stretch gap-5">
                            <Button size={"large"} variant={"basic"} label="Continue with GitHub" IconComp={GithubLogoFill} href={"https://google.com"} />
                            <Button size={"large"} variant={"secondary"} style={{ background: "#7759c2", borderColor: "#673ab7" }} label="Continue with GitLab" IconComp={GitlabLogoFill} />
                            <Button size={"large"} variant={"primary"} label="Continue with Google" IconComp={CustomGoogleIcon} />
                        </div>}
                </div>
                {searchParams.get('mode') == "email"
                    ?
                    <Button size={"large"} variant={"outline"} label="Other Login options" IconComp={ArrowLeft} href={"/login"} />
                    :
                    <Button size={"large"} variant={"outline"} label="Login with Email" IconComp={Envelope} href={"/login/?mode=email"} />}

            </div>
        </div>
        <div className="py-8  px-5 flex flex-row items-center justify-center self-stretch">
            <div className="bodyMd text-text-default">Don’t have an account?</div>
            <Button label={"Signup"} variant={"primary-plain"} size="medium" href={"/signup"} />
        </div>
    </div>
}