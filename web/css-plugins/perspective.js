function perspective() {
  // @ts-ignore
  return ({ matchUtilities }) => {
    matchUtilities({
      // @ts-ignore
      perspective: (value) => ({
        perspective: value,
      }),
    });
  };
}

export default perspective;
