import classNames from "classnames";
import { BrandLogo } from "../../../../components/branding/brand-logo";
import { Button } from "../../../../components/atoms/button";
import { ArrowLeft, ArrowRight, Envelope, EnvelopeFill, GithubLogoFill, GitlabLogoFill, GoogleLogo } from "@jengaicons/react";
import { useSearchParams } from "react-router-dom";
import { PasswordInput, TextInput } from "../../../../components/atoms/input";

export default function AuthLogin({ }) {
    const [searchParams, setSearchParams] = useSearchParams()
    return <div className={classNames("flex flex-col items-center justify-center h-full")}>
        <div className={classNames("flex flex-1 flex-col items-center self-stretch justify-center px-5 pb-8 border-b border-border-default md:py-37.5")}>
            <div className="flex flex-col items-stretch justify-center gap-8 md:w-[400px]">
                <BrandLogo darkBg={false} size={60} />
                <div className="flex flex-col items-stretch gap-8 pb-8">
                    <div className="flex flex-col gap-2 items-center md:px-12">
                        <div className={classNames("text-text-strong heading3xl text-center")}>Forgot password</div>
                        <div className="text-text-soft bodySm text-center">Enter your registered email below to receive password reset instruction.</div>
                    </div>
                    <div className="flex flex-col items-stretch gap-5">
                        <TextInput label={"Email"} placeholder={"zuko@example.com"} />
                        <Button size={"large"} variant={"primary"} label="Send instructions" DisclosureComp={ArrowRight} />
                    </div>
                </div>
            </div>
        </div>
        <div className="py-8  px-5 flex flex-row items-center justify-center self-stretch">
            <div className="bodyMd text-text-default">Remember password?</div>
            <Button label={"Login"} variant={"primary-plain"} size="medium" href={"/login"} />
        </div>
    </div>
}