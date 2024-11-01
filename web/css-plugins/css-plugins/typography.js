function typography() {
  // @ts-ignore
  return ({ addComponents }) => {
    addComponents({
      '.bodyXs': {
        '@apply kl-text-xxs kl-leading-xxs kl-font-sans kl-font-normal': {},
      },
      '.bodyMono': {
        '@apply kl-text-sm kl-leading-sm kl-font-mono': {},
      },
      '.bodyMono-medium': {
        '@apply kl-bodyMono !kl-font-medium': {},
      },
      '.bodySm': {
        '@apply kl-text-xs kl-leading-xs kl-font-sans kl-font-normal': {},
      },
      '.bodySm-medium': {
        '@apply kl-bodySm !kl-font-medium': {},
      },
      '.bodySm-semibold': {
        '@apply kl-bodySm !kl-font-semibold': {},
      },
      '.bodyMd': {
        '@apply kl-text-sm kl-leading-sm kl-font-sans': {},
      },
      '.bodyMd-medium': {
        '@apply kl-bodyMd !kl-font-medium': {},
      },
      '.bodyMd-semibold': {
        '@apply kl-bodyMd !kl-font-semibold': {},
      },
      '.bodyMd-underline': {
        '@apply kl-bodyMd kl-underline': {},
      },
      '.bodyLg': {
        '@apply kl-text-md kl-leading-md kl-font-sans': {},
      },
      '.bodyLg-medium': {
        '@apply kl-bodyLg !kl-font-medium': {},
      },
      '.bodyLg-semibold': {
        '@apply kl-bodyLg !kl-font-semibold': {},
      },
      '.bodyLg-underline': {
        '@apply kl-bodyLg kl-underline': {},
      },
      '.bodyXl': {
        '@apply kl-font-normal kl-font-sans kl-text-lg kl-leading-lg': {},
      },
      '.bodyXXl': {
        '@apply kl-font-normal kl-font-sans kl-text-xl kl-leading-bodyXXl-lineHeight':
          {},
      },
      '.bodyXl-medium': {
        '@apply kl-font-medium kl-font-sans kl-text-lg kl-leading-lg': {},
      },
      '.headingXs': {
        '@apply kl-font-semibold kl-text-xs kl-leading-xs kl-font-sans': {},
      },
      '.headingSm': {
        '@apply kl-font-semibold kl-text-sm kl-leading-sm kl-font-sans': {},
      },
      '.headingMd': {
        '@apply kl-font-semibold kl-text-md kl-leading-md kl-font-sans': {},
      },
      '.headingMd-marketing': {
        '@apply kl-headingMd !kl-font-familjen': {},
      },
      '.headingLg': {
        '@apply kl-font-semibold kl-text-lg kl-leading-md kl-font-sans': {},
      },
      '.headingLg-marketing': {
        '@apply kl-headingLg !kl-font-familjen': {},
      },
      '.headingXl': {
        '@apply kl-font-semibold kl-text-xl kl-leading-lg kl-font-sans': {},
      },
      '.headingXl-marketing': {
        '@apply kl-headingXl !kl-font-familjen': {},
      },
      '.heading2xl': {
        '@apply kl-font-semibold kl-text-2xl kl-leading-xl kl-font-sans': {},
      },
      '.heading2xl-marketing': {
        '@apply kl-heading2xl kl-font-familjen': {},
      },
      '.heading3xl': {
        '@apply kl-font-semibold kl-text-3xl kl-leading-2xl kl-font-sans': {},
      },
      '.heading3xl-marketing': {
        '@apply kl-heading3xl !kl-font-familjen': {},
      },
      '.heading3xl-1-marketing': {
        '@apply kl-heading3xl !kl-font-sriracha': {},
      },
      '.heading4xl': {
        '@apply kl-font-bold kl-text-4xl kl-leading-3xl kl-font-sans': {},
      },
      '.heading4xl-marketing': {
        '@apply kl-heading4xl !kl-font-familjen': {},
      },
      '.heading5xl-marketing': {
        '@apply kl-font-bold kl-text-5xl kl-leading-5xl kl-font-familjen': {},
      },
      '.heading5xl-1-marketing': {
        '@apply kl-font-normal kl-text-5xl kl-leading-5xl kl-font-sriracha': {},
      },
      '.heading6xl-marketing': {
        '@apply kl-font-bold kl-text-6xl kl-leading-6xl kl-font-familjen': {},
      },
      '.sriracha5xl': {
        '@apply kl-font-normal kl-text-5xl kl-leading-4xl kl-font-sriracha': {},
      },
      '.sriracha4xl': {
        '@apply kl-font-normal kl-text-4xl kl-leading-2xl-1 kl-font-sriracha':
          {},
      },
      '.sriracha3xl': {
        '@apply kl-font-normal kl-text-3xl kl-leading-xl-1 kl-font-sriracha':
          {},
      },
    });
  };
}

export default typography;
