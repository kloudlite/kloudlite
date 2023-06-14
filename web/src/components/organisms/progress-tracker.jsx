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
            <div className={classNames("w-[10px] h-[10px] rounded-full flex items-center justify-center")}>
                <div className={classNames("rounded-full",
                    {
                        "bg-icon-primary w-[10px] h-[10px]": active,
                        "bg-icon-disabled w-[5px] h-[5px]": !active
                    })}></div>
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
                {index != (items.length - 1) && <div className="flex items-center justify-center w-[10px]">
                    <svg width="10" height="35" className="-mt-[13px] -mb-[22px]">
                        <line x1="5" y1="1" x2="5" y2="34" stroke={trackerLineColor} strokeWidth="1" strokeLinecap="round" strokeDasharray="3, 4"></line>
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
        },
        {
            label: "Item 2",
            active: true,
        },
        {
            label: "Item 3",
            active: false,
        },
        {
            label: "Item 4",
            active: false,
        },
        {
            label: "Item 5",
            active: false,
        },
    ]
}