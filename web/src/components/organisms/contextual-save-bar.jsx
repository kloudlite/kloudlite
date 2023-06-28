import classNames from "classnames"
import { Button } from "../atoms/button"
import { BrandLogo } from "../branding/brand-logo"

export const ContextualSaveBar = ({ logo, logoWidth, message, saveAction, discardAction, fixed }) => {
    return <div className={classNames("bg-surface-secondary-pressed py-3 px-5",
        {
            "fixed top-0 left-0 right-0": fixed
        })}>
        <div className="flex flex-row items-center justify-between  m-auto">
            {logo && <div className="hidden md:block lg:block xl:block" width={logoWidth || 124} >
                {logo}
            </div>}
            {message && <div className="headingMd text-text-on-primary font-sans-serif">{message}</div>}
            {logo && <>
                <div></div>
                <div></div>
                <div></div>
            </>}
            <div className="gap-x-2 flex flex-row items-center">
                {discardAction && <Button label="Discard" onClick={discardAction} variant={'secondary-outline'} />}
                {saveAction && <Button label="Save" onClick={saveAction} variant={'basic'} />}
            </div>
        </div>
    </div>
}




ContextualSaveBar.defaultProps = {
    imageWidth: 124,
    logo: <BrandLogo detailed darkBg size={20} />,
    message: "Unsaved changes",
    saveAction: (e) => { console.log(e) },
    discardAction: (e) => { }
}