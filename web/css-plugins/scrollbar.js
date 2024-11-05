function scrollbar() {
  // @ts-ignore
  return ({ addBase }) => {
    addBase({
      '*': {
        '&::-webkit-scrollbar': {
          width: '6px',
          height: '6px',
        },
        '&::-webkit-scrollbar-track': {
          '@apply bg-transparent': {},
        },
        '&::-webkit-scrollbar-thumb': {
          '@apply bg-surface-basic-hovered rounded cursor-pointer': {},
        },
        '&::-webkit-scrollbar-track-piece': {
          '@apply bg-transparent': {},
        },
        '&::-webkit-scrollbar-thumb:hover': {
          '@apply bg-surface-basic-pressed': {},
        },
      },
    });
  };
}

export default scrollbar;
