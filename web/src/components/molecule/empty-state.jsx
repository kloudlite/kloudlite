import classNames from "classnames"
import { Button } from "../atoms/button"
import PropTypes from 'prop-types';

export const EmptyState = ({ image, heading, children, footer, action, secondaryAction }) => {
    return (
        <div className="flex flex-col items-center shadow-card border border-border-default rounded">
            <div className={classNames("flex flex-col items-center pt-20 pb-5 px-5")}>
                {image ?
                    <img src={image} className="max-h-43 max-w-37" />
                    : <div className="h-43 w-37 bg-surface-hovered"></div>}
                <div className="flex flex-col gap-4 pb-15">
                    <div className="headingLg text-center">{heading}</div>
                    {children && <div className="text-text-strong bodyMd text-center">
                        {children}
                    </div>}
                    {(action || secondaryAction) && <div className="flex flex-row items-center justify-center gap-2">
                        {secondaryAction && <Button content={secondaryAction?.title} variant={"outline"} onClick={secondaryAction?.click} />}
                        {action && <Button content={action?.title} variant={"primary"} onClick={action?.click} />}
                    </div>}
                    {footer && <div className="bodySm text-text-soft">{footer}</div>}
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