'use client'

import { motion, useInView } from 'motion/react'
import { useRef, type ReactNode } from 'react'
import { useReducedMotion } from './use-reduced-motion'
import { fadeInUpVariants } from './animation-variants'

interface AnimatedFeatureCardProps {
  children: ReactNode
  delay?: number
}

export function AnimatedFeatureCard({ children, delay = 0 }: AnimatedFeatureCardProps) {
  const ref = useRef(null)
  const isInView = useInView(ref, { once: true, margin: '-50px' })
  const prefersReducedMotion = useReducedMotion()

  const shouldAnimate = !prefersReducedMotion

  return (
    <motion.div
      ref={ref}
      initial={shouldAnimate ? 'hidden' : 'visible'}
      animate={isInView ? 'visible' : 'hidden'}
      variants={fadeInUpVariants}
      transition={{ delay }}
    >
      {children}
    </motion.div>
  )
}
