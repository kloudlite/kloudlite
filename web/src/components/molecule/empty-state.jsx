import classNames from "classnames"
import { Button } from "../atoms/button"
import PropTypes from 'prop-types';

export const EmptyState = ({ image, heading, children, footer, action, secondaryAction, fullwidth }) => {
    return (
        <div className="flex flex-col items-center shadow-card border border-border-default rounded">
            <div className={classNames("flex flex-col items-center",
                {
                    "max-w-[400px]": !fullwidth
                })}>
                {image ?
                    <img src={image} className="max-h-[172px] max-w-[148px] mt-[60px]" />
                    : <div className="h-[172px] w-[148px] bg-surface-hovered mt-[60px] md:mt-[67px]"></div>}
                <div className="headingLg mt-[27px] text-center">{heading}</div>
                {children && <div className="text-text-strong bodyMd mt-4 text-center">
                    {children}
                </div>}
                {(action || secondaryAction) && <div className="mt-6 flex flex-row items-center justify-center gap-2">
                    {secondaryAction && <Button label={secondaryAction?.title} style={"outline"} onClick={secondaryAction?.click} />}
                    {action && <Button label={action?.title} style={"primary"} onClick={action?.click} />}
                </div>}
                <div className={classNames("mb-20 text-center",
                    {
                        "mt-3": footer
                    })}>
                    <div className="bodySm text-text-soft">{footer}</div>
                </div>
            </div>
        </div>
    )
}





EmptyState.propTypes = {
    image: PropTypes.string,
    heading: PropTypes.string,
    children: PropTypes.any,
    footer: PropTypes.any,
    action: PropTypes.shape({
        title: PropTypes.string.isRequired,
        click: PropTypes.func
    }),
    secondaryAction: PropTypes.shape({
        title: PropTypes.string.isRequired,
        click: PropTypes.func
    }),
    fullwidth: PropTypes.bool
}


EmptyState.defaultProps = {
    heading: "This is where you’ll manage your projects",
    children: <div>You can create a new project and manage the listed project.</div>,
    fullwidth: true
}