import { DotsThreeCircleFill, DotsThreeCircleVerticalFill, DotsThreeVerticalFill } from "@jengaicons/react"
import { useGridList } from "@react-aria/gridlist"
import { useGridListItem } from "@react-aria/gridlist"
import { useFocusRing } from "@react-aria/focus"
import classNames from "classnames"
import { forwardRef, useRef } from "react"
import { Item, useListState } from "react-stately"
import { IconButton } from "~/components/atoms/button"
import { Thumbnail } from "~/components/atoms/thumbnail"

const List = (props) => {
    let state = useListState(props);
    let ref = useRef();
    let { gridProps } = useGridList(props, state, ref);
    return (
        <ul {...gridProps} ref={ref} className={classNames("flex rounded",
            {
                "flex-row flex-wrap gap-6xl ": props.mode === "grid",
                "shadow-base border-border-default flex-col": props.mode === "list"
            })}>
            {[...state.collection].map((item) => (
                <ListItem key={item.key} item={item} state={state} mode={props.mode} />
            ))}
        </ul>
    );
}










const ListItem = ({ item, state, mode }) => {
    let ref = useRef(null);
    let { rowProps, gridCellProps, isPressed } = useGridListItem(
        { node: item },
        state,
        ref
    );

    let { isFocusVisible, focusProps } = useFocusRing();

    return (
        <li
            {...rowProps}
            {...focusProps}
            ref={ref}
            className={classNames("outline-none ring-offset-1 relative bg-surface-basic-default hover:bg-surface-basic-hovered",
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
                className={classNames("cursor-pointer flex flex-col p-3xl gap-3xl ring-offset-1")}
            >
                <div className="flex flex-row items-center justify-between gap-lg">
                    <div className="flex flex-row items-center gap-xl">
                        <Thumbnailail size={'small'} rounded src={"https://images.unsplash.com/photo-1600716051809-e997e11a5d52?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHxzZWFyY2h8NHx8c2FtcGxlfGVufDB8fDB8fA%3D%3D&auto=format&fit=crop&w=800&q=60"} />
                        <div className="flex flex-col gap-sm">
                            <div className="flex flex-row gap-md items-center">
                                <div className="headingMd text-text-default">Lobster early</div>
                                <div className="w-lg h-lg bg-icon-primary rounded-full"></div>
                            </div>
                            <div className="bodyMd text-text-soft">lobster-early-kloudlite-app</div>
                        </div>
                    </div>
                    <IconButton variant="plain" icon={DotsThreeVerticalFill} size="small" />
                </div>
                <div className="flex flex-col gap-md items-start">
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
            className={classNames("cursor-pointer flex flex-row items-center justify-between px-3xl pt-3xl pb-3xl gap-3xl")}>
            <div className="flex flex-row items-center gap-xl">
                <Thumbnail size={'small'} rounded src={"https://images.unsplash.com/photo-1600716051809-e997e11a5d52?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHxzZWFyY2h8NHx8c2FtcGxlfGVufDB8fDB8fA%3D%3D&auto=format&fit=crop&w=800&q=60"} />
                <div className="flex flex-col gap-sm">
                    <div className="flex flex-row gap-md items-center">
                        <div className="headingMd text-text-default">Lobster early</div>
                        <div className="w-lg h-lg bg-icon-primary rounded-full"></div>
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
            {/* <IconButton variant="plain" icon={DotsThreeVerticalFill} size="small" onClick={(e) => {

                console.log("hello world")
                e.preventDefault()
            }} /> */}
        </div>
    )
}



export default function ResourceList({ items = [], mode = "list" }) {
    return <List
        selectionMode="none"
        selectionBehavior="toggle"
        onAction={(key) => {
            console.log("item clicked", key);
        }}
        mode={mode}
    >
        <Item>
            <ResourceItem mode={mode} />
            <button onClick={(e) => console.log(e)}>click</button>
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