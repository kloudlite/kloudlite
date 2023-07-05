export const flattenChildren = (children) => {
  return (Array.isArray(children) ? children : [children]).reduce(
    (arr, cur) => {
      let _arr = arr;
      if (Array.isArray(cur)) {
        _arr = [...arr, ...cur];
      } else {
        _arr.push(cur);
      }
      return _arr;
    },
    []
  );
};
