import classNames from "classnames"

export const Thumbnail = ({src, size, rounded}) => {
  return (<div className={
    classNames("rounded border border-border-default overflow-clip", {
      "w-8 h-8": size === "extra-small",
      "w-10 h-10": size === "small",
      "w-16 h-16": size === "medium",
      "w-20 h-20": size === "large",
  }, {
    "rounded-full": rounded,
    "rounded-md": !rounded,
  })}>
    <img src={src} className="w-full h-full object-cover" />
  </div>)
}