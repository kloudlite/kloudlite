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
          '@apply kl-bg-transparent': {},
        },
        '&::-webkit-scrollbar-thumb': {
          '@apply kl-bg-surface-basic-hovered kl-rounded kl-cursor-pointer': {},
        },
        '&::-webkit-scrollbar-track-piece': {
          '@apply kl-bg-transparent': {},
        },
        '&::-webkit-scrollbar-thumb:hover': {
          '@apply kl-bg-surface-basic-pressed': {},
        },
      },
    });
  };
}

export default scrollbar;
