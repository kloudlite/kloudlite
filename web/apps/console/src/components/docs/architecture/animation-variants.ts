import type { Variants } from 'motion/react'
import { ANIMATION_DURATION, ANIMATION_EASING, STAGGER_DELAY } from './animation-config'

/**
 * Shared animation variants for consistent animations across components
 */

export const fadeInVariants: Variants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      duration: ANIMATION_DURATION.normal,
      ease: ANIMATION_EASING.easeOut,
    },
  },
}

export const fadeInUpVariants: Variants = {
  hidden: { opacity: 0, y: 20 },
  visible: {
    opacity: 1,
    y: 0,
    transition: {
      duration: ANIMATION_DURATION.normal,
      ease: ANIMATION_EASING.easeOut,
    },
  },
}

export const scaleInVariants: Variants = {
  hidden: { opacity: 0, scale: 0.8 },
  visible: {
    opacity: 1,
    scale: 1,
    transition: {
      duration: ANIMATION_DURATION.normal,
      ease: ANIMATION_EASING.easeOut,
    },
  },
}

export const staggerContainerVariants: Variants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      staggerChildren: STAGGER_DELAY.slow,
      delayChildren: 0.1,
    },
  },
}

export const slideInLeftVariants: Variants = {
  hidden: { opacity: 0, x: -20 },
  visible: {
    opacity: 1,
    x: 0,
    transition: {
      duration: ANIMATION_DURATION.normal,
      ease: ANIMATION_EASING.easeOut,
    },
  },
}
