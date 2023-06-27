import { DotsThreeCircleFill, DotsThreeCircleVerticalFill, DotsThreeVerticalFill } from "@jengaicons/react"
import { IconButton } from "../../../../components/atoms/button"
import { Thumbnail } from "../../../../components/atoms/thumbnail"
import classNames from "classnames"
import { forwardRef, useRef } from "react"
import { mergeProps, useButton, useFocusRing, useGridList, useGridListItem } from "react-aria"
import { Item, useListState } from "react-stately"


const AriaButton = forwardRef(({ className, ...props }, ref) => {
    let { buttonProps } = useButton(props, ref);
    return <button {...buttonProps} ref={ref} className={className}>{props.children}</button>;
})

function List(props) {
    let state = useListState(props);
    let ref = useRef();
    let { gridProps } = useGridList(props, state, ref);

    console.log(props);

    return (
        <ul {...gridProps} ref={ref} className={classNames("flex rounded",
            {
                "flex-row flex-wrap gap-10 ": props.mode === "grid",
                "shadow-base border-border-default flex-col": props.mode === "list"
            })}>
            {[...state.collection].map((item) => (
                <ListItem key={item.key} item={item} state={state} mode={props.mode} />
            ))}
        </ul>
    );
}

function ListItem({ item, state, mode }) {
    let ref = useRef(null);
    let { rowProps, gridCellProps, isPressed } = useGridListItem(
        { node: item },
        state,
        ref
    );

    let { isFocusVisible, focusProps } = useFocusRing();

    return (
        <li
            {...mergeProps(rowProps, focusProps)}
            ref={ref}
            className={classNames("outline-none ring-offset-1 relative bg-surface-default hover:bg-surface-hovered",
                {
                    "focus-visible:ring-2 focus:ring-border-focus z-10 ring-offset-0 border-surface-default": isFocusVisible,
                    "border border-border-default rounded w-92 shadow-base": mode === "grid",
                    "border-b border-border-disabled first:rounded-t last:rounded-b": mode === "list"
                })}
        >
            <div {...gridCellProps}>
                {item.rendered}
            </div>
        </li>
    );
}

export const ResourceItem = ({ mode = "list" }) => {
    if (mode === "grid")
        return (
            <div
                className={classNames("cursor-pointer flex flex-col  p-4.75 gap-5 ring-offset-1")}
            >
                <div className="flex flex-row items-center justify-between gap-2">
                    <div className="flex flex-row items-center gap-3">
                        <Thumbnail size={'small'} rounded src={"https://images.unsplash.com/photo-1600716051809-e997e11a5d52?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHxzZWFyY2h8NHx8c2FtcGxlfGVufDB8fDB8fA%3D%3D&auto=format&fit=crop&w=800&q=60"} />
                        <div className="flex flex-col gap-0.5">
                            <div className="flex flex-row gap-1 items-center">
                                <div className="headingMd text-text-default">Lobster early</div>
                                <div className="w-2 h-2 bg-icon-primary rounded-full"></div>
                            </div>
                            <div className="bodyMd text-text-soft">lobster-early-kloudlite-app</div>
                        </div>
                    </div>
                    <IconButton variant="plain" IconComp={DotsThreeVerticalFill} size="small" />
                </div>
                <div className="flex flex-col gap-1 items-start">
                    <div className="bodyMd text-text-strong">dusty-crossbow.com/projects</div>
                    <div className="bodyMd text-text-strong">Plaxonic</div>
                </div>
                <div className="flex flex-col items-start">
                    <div className="bodyMd text-text-strong">Reyan updated the project</div>
                    <div className="bodyMd text-text-soft">3 days ago</div>
                </div>
            </div>
        )
    return (
        <div
            className={classNames("cursor-pointer flex flex-row items-center justify-between px-5 pt-5 pb-4.75 gap-5")}>
            <div className="flex flex-row items-center gap-3">
                <Thumbnail size={'small'} rounded src={"https://images.unsplash.com/photo-1600716051809-e997e11a5d52?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHxzZWFyY2h8NHx8c2FtcGxlfGVufDB8fDB8fA%3D%3D&auto=format&fit=crop&w=800&q=60"} />
                <div className="flex flex-col gap-0.5">
                    <div className="flex flex-row gap-1 items-center">
                        <div className="headingMd text-text-default">Lobster early</div>
                        <div className="w-2 h-2 bg-icon-primary rounded-full"></div>
                    </div>
                    <div className="bodyMd text-text-soft">lobster-early-kloudlite-app</div>
                </div>
            </div>
            <div className="bodyMd text-text-strong">dusty-crossbow.com/projects</div>
            <div className="bodyMd text-text-strong">Plaxonic</div>
            <div className="flex flex-col">
                <div className="bodyMd text-text-strong">Reyan updated the project</div>
                <div className="bodyMd text-text-soft">3 days ago</div>
            </div>
            <IconButton variant="plain" IconComp={DotsThreeVerticalFill} size="small" onClick={(e) => { console.log("hello world") }} />
        </div>
    )
}



export default function ResourceList({ items = [], mode = "list" }) {
    return <List
        selectionMode="none"
        selectionBehavior="toggle"
        onAction={(key) => alert(`Opening item ${key}...`)}
        mode={mode}
    >
        <Item>
            <ResourceItem mode={mode} />
        </Item>
        <Item>
            <ResourceItem mode={mode} />
        </Item>
        <Item>
            <ResourceItem mode={mode} />
        </Item>
        <Item>
            <ResourceItem mode={mode} />
        </Item>
    </List>
}