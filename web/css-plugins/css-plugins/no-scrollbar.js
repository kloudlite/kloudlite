function noScrollbar() {
  // @ts-ignore
  return ({ addComponents }) => {
    addComponents({
      '.no-scrollbar-base::-webkit-scrollbar': {
        display: 'none',
      },
      '.no-scrollbar-base': {
        '-ms-overflow-style': 'none',
        'scrollbar-width': 'none',
      },
      '.no-scrollbar': {
        '@apply kl-no-scrollbar-base': {},
      },
    });
  };
}

export default noScrollbar;
