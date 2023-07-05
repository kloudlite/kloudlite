import classNames from "classnames";
import { Button } from "~/root/src/components/atoms/button";
import { ArrowLeft, Envelope, EnvelopeFill, GithubLogoFill, GitlabLogoFill, GoogleLogo } from "@jengaicons/react";
import { useSearchParams, Link } from "@remix-run/react";
import { PasswordInput, TextInput } from "~/root/src/components/atoms/input";
import { BrandLogo } from "~/root/src/components/branding/brand-logo";

const CustomGoogleIcon = (props) => {
    return <GoogleLogo {...props} weight={4}></GoogleLogo>
}

export default () => {
    const [searchParams, setSearchParams] = useSearchParams()
    return <div className={classNames("flex flex-col items-center justify-center h-full")}>
        <div className={classNames("flex flex-1 flex-col items-center self-stretch justify-center px-5 pb-8 border-b border-border-default md:py-37.5")}>
            <div className="flex flex-col items-stretch justify-center gap-8 md:w-[400px]">
                <BrandLogo darkBg={false} size={60} />
                <form className="flex flex-col items-stretch gap-8 border-b pb-8 border-border-default" >
                    <div className="flex flex-col gap-2 items-center md:px-12">
                        <div className={classNames("text-text-strong heading3xl text-center")}>Login to Kloudlite</div>
                        <div className="text-text-soft bodySm text-center">To access your DevOps console, Please provide your login credentials.</div>
                    </div>
                    {searchParams.get('mode') == "email"
                        ?
                        <div className="flex flex-col items-stretch gap-5">
                            <TextInput label={"Email"} placeholder={"ex: john@company.com"} />
                            <PasswordInput label={"Password"} placeholder={"XXXXXX"} extra={<Button size={"medium"} variant={"primary-plain"} content={"Forgot password"} href={"/forgotpassword"} LinkComponent={Link} />} />
                            <Button size={"large"} variant={"primary"} content={<span className="bodyLg-medium">Continue with Email</span>} IconComp={EnvelopeFill} block type="submit" />
                        </div>
                        :
                        <div className="flex flex-col items-stretch gap-5">
                            <Button size={"large"} variant={"basic"} content={<span className="bodyLg-medium">Continue with GitHub</span>} IconComp={GithubLogoFill} href={"https://google.com"} block LinkComponent={Link} />
                            <Button size={"large"} variant={"secondary"} style={{ background: "#7759c2", borderColor: "#673ab7" }} content={<span className="bodyLg-medium">Continue with GitLab</span>} IconComp={GitlabLogoFill} block LinkComponent={Link} />
                            <Button size={"large"} variant={"primary"} content={<span className="bodyLg-medium">Continue with Google</span>} IconComp={CustomGoogleIcon} block LinkComponent={Link} />
                        </div>}
                </form>
                {searchParams.get('mode') == "email"
                    ?
                    <Button size={"large"} variant={"outline"} content={<span className="bodyLg-medium">Other Login options</span>} IconComp={ArrowLeft} href={"/login"} block LinkComponent={Link} />
                    :
                    <Button size={"large"} variant={"outline"} content={<span className="bodyLg-medium">Login with Email</span>} IconComp={Envelope} href={"/login/?mode=email"} block LinkComponent={Link} />}

            </div>
        </div>
        <div className="py-8  px-5 flex flex-row items-center justify-center self-stretch">
            <div className="bodyMd text-text-default">Don’t have an account?</div>
            <Button content={"Signup"} variant={"primary-plain"} size="medium" href={"/signup"} LinkComponent={Link} />
        </div>
    </div>
}