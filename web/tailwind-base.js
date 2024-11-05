import noScrollbar from './css-plugins/no-scrollbar.js';
import noSpinner from './css-plugins/no-spinner.js';
import scrollbar from './css-plugins/scrollbar.js';
import typography from './css-plugins/typography.js';
import perspective from './css-plugins/perspective.js';

const primitives = {
  spacing: {
    '025': '1px',
    '05': '2px',
    1: '4px',
    2: '8px',
    3: '12px',
    4: '16px',
    5: '20px',
    6: '24px',
    8: '32px',
    10: '40px',
    12: '48px',
    16: '64px',
    20: '80px',
    24: '96px',
    32: '128px',
    40: '160px',
    48: '192px',
    60: '240px',
    64: '256px',
    'form-text-field-height': '36px',
  },
  fontSize: {
    xxs: '10px',
    xs: '12px',
    sm: '14px',
    md: '16px',
    lg: '20px',
    xl: '24px',
    '2xl': '28px',
    '3xl': '32px',
    '4xl': '40px',
    '5xl': '56px',
    '6xl': '64px',
    '7xl': '72px',
  },
  lineHeight: {
    xxs: '14px',
    xs: '16px',
    sm: '20px',
    md: '24px',
    lg: '28px',
    xl: '32px',
    'xl-1': '38px',
    '2xl': '40px',
    '2xl-1': '46px',
    '3xl': '48px',
    '4xl': '60px',
    '5xl': '64px',
    '6xl': '76px',
    '7xl': '84px',
    'bodyXXl-lineHeight': '36px',
  },
};

const width = {
  '8xl': '90rem',
};

const config = {
  darkMode: 'class',
  theme: {
    extend: {
      keyframes: {
        animation: {
          'spin-slow': 'indeterminate 1s infinite linear',
        },
        indeterminate: {
          '40%': { transform: 'translateX(0) scaleX(0.4)' },
          '100%': { transform: 'translateX(100%) scaleX(0.5)' },
        },
        enterFromRight: {
          from: { opacity: '0', transform: 'translateX(200px)' },
          to: { opacity: '1', transform: 'translateX(0)' },
        },
        enterFromLeft: {
          from: { opacity: '0', transform: 'translateX(-200px)' },
          to: { opacity: '1', transform: 'translateX(0)' },
        },
        exitToRight: {
          from: { opacity: '1', transform: 'translateX(0)' },
          to: { opacity: '0', transform: 'translateX(200px)' },
        },
        exitToLeft: {
          from: { opacity: '1', transform: 'translateX(0)' },
          to: { opacity: '0', transform: 'translateX(-200px)' },
        },
        scaleIn: {
          from: { opacity: '0', transform: 'rotateX(-10deg) scale(0.9)' },
          to: { opacity: '1', transform: 'rotateX(0deg) scale(1)' },
        },
        scaleOut: {
          from: { opacity: '1', transform: 'rotateX(0deg) scale(1)' },
          to: { opacity: '0', transform: 'rotateX(-10deg) scale(0.95)' },
        },
        fadeIn: {
          from: { opacity: '0' },
          to: { opacity: '1' },
        },
        fadeOut: {
          from: { opacity: '1' },
          to: { opacity: '0' },
        },
        slideDown: {
          from: { height: '0px' },
          to: { height: 'var(--radix-accordion-content-height)' },
        },
        slideUp: {
          from: { height: 'var(--radix-accordion-content-height)' },
          to: { height: '0px' },
        },
        slideDownAndFade: {
          from: { opacity: '0', transform: 'translateY(-2px)' },
          to: { opacity: '1', transform: 'translateY(0)' },
        },
        slideLeftAndFade: {
          from: { opacity: '0', transform: 'translateX(2px)' },
          to: { opacity: '1', transform: 'translateX(0)' },
        },
        slideUpAndFade: {
          from: { opacity: '0', transform: 'translateY(-2px) scale(0.95)' },
          to: { opacity: '1', transform: 'translateY(0) scale(1)' },
        },
        slideRightAndFade: {
          from: { opacity: '0', transform: 'translateX(-2px)' },
          to: { opacity: '1', transform: 'translateX(0)' },
        },
      },
      animation: {
        scaleIn: 'scaleIn 200ms ease',
        scaleOut: 'scaleOut 200ms ease',
        fadeIn: 'fadeIn 200ms ease',
        fadeOut: 'fadeOut 200ms ease',
        enterFromLeft: 'enterFromLeft 250ms ease',
        enterFromRight: 'enterFromRight 250ms ease',
        exitToLeft: 'exitToLeft 250ms ease',
        exitToRight: 'exitToRight 250ms ease',
        slideDown: 'slideDown 300ms cubic-bezier(0.87, 0, 0.13, 1)',
        slideUp: 'slideUp 300ms cubic-bezier(0.87, 0, 0.13, 1)',
        slideDownAndFade:
          'slideDownAndFade 400ms cubic-bezier(0.16, 1, 0.3, 1)',
        slideLeftAndFade:
          'slideLeftAndFade 400ms cubic-bezier(0.16, 1, 0.3, 1)',
        slideUpAndFade: 'slideUpAndFade 400ms cubic-bezier(0.16, 1, 0.3, 1)',
        slideRightAndFade:
          'slideRightAndFade 400ms cubic-bezier(0.16, 1, 0.3, 1)',
      },
      boxShadow: {
        button: '0px 1px 4px rgba(0, 0, 0, 0.05)',
        card: [
          '0px 2px 1px rgba(0, 0, 0, 0.05)',
          '0px 0px 1px rgba(0, 0, 0, 0.25)',
        ],
        popover: [
          '0px 0px 2px rgba(0, 0, 0, 0.2)',
          '0px 2px 10px rgba(0, 0, 0, 0.1)',
        ],
        popup: [
          '0px 0px 1px rgba(0, 0, 0, 0.2)',
          '0px 26px 80px rgba(0, 0, 0, 0.2)',
        ],
        'shadow-2': [
          '0px 0px 1px  rgba(244, 244, 245, 1)',
          '0px 26px 80px rgba(250, 250, 250, 0)',
        ],
        deep: [
          '0px 0px 0px 1px rgba(6, 44, 82, 0.1)',
          '0px 2px 16px rgba(33, 43, 54, 0.08)',
        ],
        modal: [
          '0px 0px 1px rgba(0, 0, 0, 0.2)',
          '0px 26px 80px rgba(0, 0, 0, 0.2)',
        ],
        base: [
          '0px 1px 3px rgba(63, 63, 68, 0.15)',
          '0px 0px 0px 1px rgba(63, 63, 68, 0.05)',
        ],
        focus: '0px 0px 0px 2px #60A5FA',
        'darktheme-popover': [
          '0px 0px 2px 0px rgba(250, 250, 250, 0.20)',
          '0px 2px 10px 0px rgba(250, 250, 250, 0.10)',
        ],
      },
      maxWidth: { ...width },
      minWidth: { ...width },
    },
    fontFamily: {
      sans: [
        '"Inter var", sans-serif',
        {
          fontFeatureSettings: '"cv02", "cv03", "cv04", "cv11"',
        },
      ],
      mono: ['Roboto Mono', 'monospace'],
      familjen: ['Familjen Grotesk', 'sans-serif'],
      sriracha: ['Sriracha', 'cursive'],
    },
    screens: {
      sm: '490px',
      smMd: '640px',
      md: '768px',
      lg: '1024px',
      xl: '1280px',
      '2xl': '1440px',
      '2xl-md': '1536px',
      '3xl': '1920px',
    },
    fontSize: { ...primitives.fontSize },
    lineHeight: { ...primitives.lineHeight },
    spacing: {
      0: '0px',
      xs: primitives.spacing['025'],
      sm: primitives.spacing['05'],
      md: primitives.spacing['1'],
      lg: primitives.spacing['2'],
      xl: primitives.spacing['3'],
      '2xl': primitives.spacing['4'],
      '3xl': primitives.spacing['5'],
      '4xl': primitives.spacing['6'],
      '5xl': primitives.spacing['8'],
      '6xl': primitives.spacing['10'],
      '7xl': primitives.spacing['12'],
      '8xl': primitives.spacing['16'],
      '9xl': primitives.spacing['20'],
      '10xl': primitives.spacing['24'],
      '11xl': primitives.spacing['32'],
      '12xl': primitives.spacing['40'],
      '13xl': primitives.spacing['48'],
      '14xl': primitives.spacing['60'],
      '15xl': primitives.spacing['64'],
      'form-text-field-height': primitives.spacing['form-text-field-height'],
    },
    colors: {
      surface: {
        basic: {
          default:
            'color-mix(in srgb, var(--surface-basic-default) calc(100% * <alpha-value>), transparent)',
          subdued:
            'color-mix(in srgb, var(--surface-basic-subdued) calc(100% * <alpha-value>), transparent)',
          hovered: 'var(--surface-basic-hovered)',
          pressed: 'var(--surface-basic-pressed)',
          input: 'var(--surface-basic-input)',
          active: 'var(--surface-basic-active)',
          'overlay-bg':
            'color-mix(in srgb, var(--surface-basic-overlay-bg) calc(100% * <alpha-value>), transparent)',
          'container-bg':
            'color-mix(in srgb, var(--surface-basic-container-bg) calc(100% * <alpha-value>), transparent)',
        },
        primary: {
          default: 'var(--surface-primary-default)',
          subdued: 'var(--surface-primary-subdued)',
          hovered: 'var(--surface-primary-hovered)',
          pressed: 'var(--surface-primary-pressed)',
          selected: 'var(--surface-primary-selected)',
        },
        secondary: {
          default: 'var(--surface-secondary-default)',
          subdued: 'var(--surface-secondary-subdued)',
          hovered: 'var(--surface-secondary-hovered)',
          pressed: 'var(--surface-secondary-pressed)',
        },
        tertiary: {
          default: 'var(--surface-tertiary-default)',
          hovered: 'var(--surface-tertiary-hovered)',
          pressed: 'var(--surface-tertiary-pressed)',
          active: 'var(--surface-tertiary-active)',
        },
        critical: {
          default: 'var(--surface-critical-default)',
          subdued: 'var(--surface-critical-subdued)',
          hovered: 'var(--surface-critical-hovered)',
          pressed: 'var(--surface-critical-pressed)',
        },
        warning: {
          default: 'var(--surface-warning-default)',
          subdued: 'var(--surface-warning-subdued)',
          hovered: 'var(--surface-warning-hovered)',
          pressed: 'var(--surface-warning-pressed)',
        },
        success: {
          default: 'var(--surface-success-default)',
          subdued: 'var(--surface-success-subdued)',
          hovered: 'var(--surface-success-hovered)',
          pressed: 'var(--surface-success-pressed)',
        },
        purple: {
          default: 'var(--surface-purple-default)',
          hovered: 'var(--surface-purple-hovered)',
          pressed: 'var(--surface-purple-pressed)',
        },
      },
      text: {
        default:
          'color-mix(in srgb, var(--text-default) calc(100% * <alpha-value>), transparent)',
        soft: 'var(--text-soft)',
        strong: 'var(--text-strong)',
        disabled: 'var(--text-disabled)',
        primary: 'var(--text-primary)',
        'on-primary': 'var(--text-on-primary)',
        'on-secondary': 'var(--text-on-secondary)',
        secondary: 'var(--text-secondary)',
        critical: 'var(--text-critical)',
        warning: 'var(--text-warning)',
        success: 'var(--text-success)',
      },
      icon: {
        default: 'var(--icon-default)',
        soft: 'var(--icon-soft)',
        strong: 'var(--icon-strong)',
        disabled: 'var(--icon-disabled)',
        primary: 'var(--icon-primary)',
        'on-primary': 'var(--icon-on-primary)',
        'on-secondary': 'var(--icon-on-secondary)',
        secondary: 'var(--icon-secondary)',
        critical: 'var(--icon-critical)',
        warning: 'var(--icon-warning)',
        success: 'var(--icon-success)',
        logo: 'var(--icon-logo)',
      },
      border: {
        default: 'var(--border-default)',
        dark: 'var(--border-dark)',
        disabled: 'var(--border-disabled)',
        primary: 'var(--border-primary)',
        focus: 'var(--border-focus)',
        secondary: 'var(--border-secondary)',
        tertiary: 'var(--border-tertiary)',
        critical: 'var(--border-critical)',
        warning: 'var(--border-warning)',
        success: 'var(--border-success)',
        purple: 'var(--border-purple)',
      },
      transparent: 'transparent',
      white: 'white',
      black: 'black',
      dark: {},
    },
  },
  plugins: [
    typography(),
    scrollbar(),
    noScrollbar(),
    noSpinner(),
    perspective(),
  ],
};

export const LightTitlebarColor = config.theme.colors.surface.basic.subdued;
export const ChipGroupPaddingTop = config.theme.spacing.xl;

export default config;
