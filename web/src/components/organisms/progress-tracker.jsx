import classNames from "classnames"
import PropTypes from "prop-types"


const trackerLineColor = "#D4D4D8"

const ProgressTrackerItem = ({ label, active }) => {
    return (
        <div className={classNames("flex flex-row gap-x-3 items-center",
            {
                "bodyMd-semibold text-text-default": active,
                "bodyMd text-text-disabled": !active
            })}>
            <div className={classNames("w-2.5 h-2.5 rounded-full flex items-center justify-center")}>
                {active && <svg width="10" height="10" viewBox="0 0 10 10" xmlns="http://www.w3.org/2000/svg" className="fill-icon-primary">
                    <circle cx="5" cy="5" r="5" />
                </svg>}
                {!active && <svg width="10" height="10" viewBox="0 0 10 10" xmlns="http://www.w3.org/2000/svg" className="fill-border-default">
                    <circle cx="5" cy="5" r="2.5" />
                </svg>}
            </div>
            <div className="py-2">{label}</div>
        </div>
    )
}

export const ProgressTracker = ({ items }) => {
    return <div className="flex flex-col gap-y-2">
        {items && items.map((item, index) => {
            return <div className="flex flex-col" key={item.key}>
                <ProgressTrackerItem active={item.active} label={item.label} />
                {index != (items.length - 1) && <div className="flex items-center justify-center w-2.5">
                    <svg width="10" height="35" className="-mt-3.25 -mb-5.5 stroke-border-default">
                        <line x1="5" y1="1" x2="5" y2="34" strokeWidth="1" strokeLinecap="round" strokeDasharray="3, 4"></line>
                    </svg>

                </div>}
            </div>
        })}
    </div>
}


ProgressTracker.propTypes = {
    items: PropTypes.arrayOf(PropTypes.shape({
        label: PropTypes.string.isRequired,
        active: PropTypes.bool,
        key: PropTypes.any
    })).isRequired
}


ProgressTracker.defaultProps = {
    items: [
        {
            label: "Item 1",
            active: true,
            key: 1,
        },
        {
            label: "Item 2",
            active: true,
            key: 2,
        },
        {
            label: "Item 3",
            active: false,
            key: 3,
        },
        {
            label: "Item 4",
            active: false,
            key: 4,
        },
        {
            label: "Item 5",
            active: false,
            key: 5
        },
    ]
}