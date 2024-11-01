function typography() {
  // @ts-ignore
  return ({ addComponents }) => {
    addComponents({
      '.bodyXs': {
        '@apply text-xxs leading-xxs font-sans font-normal': {},
      },
      '.bodyMono': {
        '@apply text-sm leading-sm font-mono': {},
      },
      '.bodyMono-medium': {
        '@apply bodyMono !font-medium': {},
      },
      '.bodySm': {
        '@apply text-xs leading-xs font-sans font-normal': {},
      },
      '.bodySm-medium': {
        '@apply bodySm !font-medium': {},
      },
      '.bodySm-semibold': {
        '@apply bodySm !font-semibold': {},
      },
      '.bodyMd': {
        '@apply text-sm leading-sm font-sans': {},
      },
      '.bodyMd-medium': {
        '@apply bodyMd !font-medium': {},
      },
      '.bodyMd-semibold': {
        '@apply bodyMd !font-semibold': {},
      },
      '.bodyMd-underline': {
        '@apply bodyMd underline': {},
      },
      '.bodyLg': {
        '@apply text-md leading-md font-sans': {},
      },
      '.bodyLg-medium': {
        '@apply bodyLg !font-medium': {},
      },
      '.bodyLg-semibold': {
        '@apply bodyLg !font-semibold': {},
      },
      '.bodyLg-underline': {
        '@apply bodyLg underline': {},
      },
      '.bodyXl': {
        '@apply font-normal font-sans text-lg leading-lg': {},
      },
      '.bodyXXl': {
        '@apply font-normal font-sans text-xl leading-bodyXXl-lineHeight': {},
      },
      '.bodyXl-medium': {
        '@apply font-medium font-sans text-lg leading-lg': {},
      },
      '.headingXs': {
        '@apply font-semibold text-xs leading-xs font-sans': {},
      },
      '.headingSm': {
        '@apply font-semibold text-sm leading-sm font-sans': {},
      },
      '.headingMd': {
        '@apply font-semibold text-md leading-md font-sans': {},
      },
      '.headingMd-marketing': {
        '@apply headingMd !font-familjen': {},
      },
      '.headingLg': {
        '@apply font-semibold text-lg leading-md font-sans': {},
      },
      '.headingLg-marketing': {
        '@apply headingLg !font-familjen': {},
      },
      '.headingXl': {
        '@apply font-semibold text-xl leading-lg font-sans': {},
      },
      '.headingXl-marketing': {
        '@apply headingXl !font-familjen': {},
      },
      '.heading2xl': {
        '@apply font-semibold text-2xl leading-xl font-sans': {},
      },
      '.heading2xl-marketing': {
        '@apply heading2xl font-familjen': {},
      },
      '.heading3xl': {
        '@apply font-semibold text-3xl leading-2xl font-sans': {},
      },
      '.heading3xl-marketing': {
        '@apply heading3xl !font-familjen': {},
      },
      '.heading3xl-1-marketing': {
        '@apply heading3xl !font-sriracha': {},
      },
      '.heading4xl': {
        '@apply font-bold text-4xl leading-3xl font-sans': {},
      },
      '.heading4xl-marketing': {
        '@apply heading4xl !font-familjen': {},
      },
      '.heading5xl-marketing': {
        '@apply font-bold text-5xl leading-5xl font-familjen': {},
      },
      '.heading5xl-1-marketing': {
        '@apply font-normal text-5xl leading-5xl font-sriracha': {},
      },
      '.heading6xl-marketing': {
        '@apply font-bold text-6xl leading-6xl font-familjen': {},
      },
      '.sriracha5xl': {
        '@apply font-normal text-5xl leading-4xl font-sriracha': {},
      },
      '.sriracha4xl': {
        '@apply font-normal text-4xl leading-2xl-1 font-sriracha': {},
      },
      '.sriracha3xl': {
        '@apply font-normal text-3xl leading-xl-1 font-sriracha': {},
      },
    });
  };
}

export default typography;
