'use client'

import { motion } from 'motion/react'
import { Server, Boxes, Database, Code2 } from 'lucide-react'
import { useReducedMotion } from './use-reduced-motion'
import { scaleInVariants, staggerContainerVariants, fadeInUpVariants, slideInLeftVariants } from './animation-variants'
import { ANIMATION_DURATION } from './animation-config'

export function AnimatedArchitectureDiagram() {
  const prefersReducedMotion = useReducedMotion()

  // If user prefers reduced motion, disable animations
  const shouldAnimate = !prefersReducedMotion

  return (
    <section className="mb-12 sm:mb-16">
      <div className="bg-gradient-to-br from-blue-50/50 via-purple-50/30 to-indigo-50/50 dark:from-slate-900 dark:via-slate-800 dark:to-slate-900 rounded-xl border-2 border-slate-200 dark:border-slate-700 p-8 sm:p-12 mb-6">
        <div className="flex flex-col items-center">
          {/* Control Node */}
          <motion.div
            className="relative z-10 mb-8"
            initial={shouldAnimate ? 'hidden' : 'visible'}
            animate="visible"
            variants={scaleInVariants}
          >
            <div className="bg-white dark:bg-slate-950 rounded-2xl border-3 border-blue-500 shadow-2xl px-16 py-10 relative">
              <motion.div
                className="absolute -top-3 -left-3 bg-gradient-to-r from-blue-600 to-blue-500 text-white rounded-lg px-3 py-1 text-xs font-bold uppercase tracking-wider shadow-lg"
                initial={shouldAnimate ? { opacity: 0, x: -10 } : { opacity: 1, x: 0 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: 0.3, duration: ANIMATION_DURATION.normal }}
              >
                Control Plane
              </motion.div>
              <div className="flex flex-col items-center gap-3">
                <motion.div
                  className="bg-gradient-to-br from-blue-500 to-blue-600 rounded-full p-3 shadow-lg"
                  animate={shouldAnimate ? {
                    rotate: [0, 5, -5, 0],
                  } : {}}
                  transition={{
                    delay: 0.5,
                    duration: 0.8,
                    ease: 'easeInOut',
                  }}
                >
                  <Server className="h-8 w-8 text-white" />
                </motion.div>
                <h4 className="text-slate-800 dark:text-slate-100 text-2xl font-bold m-0">
                  Control Node
                </h4>
                <p className="text-slate-600 dark:text-slate-400 text-sm m-0 font-mono">
                  {'{subdomain}'}.khost.dev
                </p>
              </div>
            </div>
          </motion.div>

          {/* Connection Lines */}
          <motion.div
            className="relative w-full max-w-4xl mb-8"
            initial={shouldAnimate ? { opacity: 0 } : { opacity: 1 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.8, duration: ANIMATION_DURATION.normal }}
          >
            <div className="flex justify-center">
              <div className="relative">
                {/* Vertical line with drawing animation */}
                <motion.div
                  className="w-1 h-16 bg-gradient-to-b from-blue-500 via-purple-500 to-purple-600 mx-auto"
                  initial={shouldAnimate ? { scaleY: 0, originY: 0 } : { scaleY: 1 }}
                  animate={{ scaleY: 1 }}
                  transition={{ delay: 1.0, duration: ANIMATION_DURATION.slow, ease: 'easeOut' }}
                />
                {/* Arrow head */}
                <motion.div
                  className="absolute -bottom-2 left-1/2 -translate-x-1/2"
                  initial={shouldAnimate ? { opacity: 0, scale: 0 } : { opacity: 1, scale: 1 }}
                  animate={{ opacity: 1, scale: 1 }}
                  transition={{ delay: 1.6, duration: ANIMATION_DURATION.fast }}
                >
                  <div className="w-0 h-0 border-l-[8px] border-l-transparent border-r-[8px] border-r-transparent border-t-[12px] border-t-purple-600"></div>
                </motion.div>
              </div>
            </div>

            {/* Horizontal branching lines */}
            <div className="relative h-12">
              <motion.div
                className="absolute top-0 left-1/2 w-1 h-12 bg-purple-600"
                initial={shouldAnimate ? { scaleY: 0, originY: 0 } : { scaleY: 1 }}
                animate={{ scaleY: 1 }}
                transition={{ delay: 1.8, duration: ANIMATION_DURATION.normal }}
              />
              <motion.div
                className="absolute top-6 left-[15%] right-[15%] h-1 bg-gradient-to-r from-emerald-500 via-purple-600 to-amber-500"
                initial={shouldAnimate ? { scaleX: 0 } : { scaleX: 1 }}
                animate={{ scaleX: 1 }}
                transition={{ delay: 2.0, duration: ANIMATION_DURATION.slow, ease: 'easeOut' }}
              />
              <motion.div
                className="absolute top-6 left-[15%] w-1 h-6 bg-emerald-500"
                initial={shouldAnimate ? { scaleY: 0, originY: 0 } : { scaleY: 1 }}
                animate={{ scaleY: 1 }}
                transition={{ delay: 2.6, duration: ANIMATION_DURATION.fast }}
              />
              <motion.div
                className="absolute top-6 right-[15%] w-1 h-6 bg-amber-500"
                initial={shouldAnimate ? { scaleY: 0, originY: 0 } : { scaleY: 1 }}
                animate={{ scaleY: 1 }}
                transition={{ delay: 2.6, duration: ANIMATION_DURATION.fast }}
              />
            </div>
          </motion.div>

          {/* Workmachines Group */}
          <motion.div
            className="relative w-full max-w-4xl"
            initial={shouldAnimate ? { opacity: 0 } : { opacity: 1 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 2.8, duration: ANIMATION_DURATION.normal }}
          >
            {/* Stack effect layers */}
            <motion.div
              className="absolute inset-0 bg-white dark:bg-slate-950 rounded-2xl border-3 border-purple-300 dark:border-purple-700 translate-x-3 translate-y-3 opacity-30"
              initial={shouldAnimate ? { opacity: 0, scale: 0.9 } : { opacity: 0.3, scale: 1 }}
              animate={{ opacity: 0.3, scale: 1 }}
              transition={{ delay: 2.8, duration: ANIMATION_DURATION.normal }}
            />
            <motion.div
              className="absolute inset-0 bg-white dark:bg-slate-950 rounded-2xl border-3 border-purple-400 dark:border-purple-600 translate-x-1.5 translate-y-1.5 opacity-50"
              initial={shouldAnimate ? { opacity: 0, scale: 0.95 } : { opacity: 0.5, scale: 1 }}
              animate={{ opacity: 0.5, scale: 1 }}
              transition={{ delay: 3.0, duration: ANIMATION_DURATION.normal }}
            />

            {/* Main Workmachine container */}
            <motion.div
              className="relative bg-white dark:bg-slate-950 rounded-2xl border-3 border-purple-500 p-8 shadow-2xl"
              initial={shouldAnimate ? { opacity: 0, y: 20 } : { opacity: 1, y: 0 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 3.2, duration: ANIMATION_DURATION.normal }}
            >
              <motion.div
                className="absolute -top-3 -left-3 bg-gradient-to-r from-purple-600 to-purple-500 text-white rounded-lg px-3 py-1 text-xs font-bold uppercase tracking-wider shadow-lg"
                initial={shouldAnimate ? { opacity: 0, x: -10 } : { opacity: 1, x: 0 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: 3.4, duration: ANIMATION_DURATION.normal }}
              >
                Compute Layer
              </motion.div>

              <div className="flex items-center gap-3 mb-8 justify-center">
                <motion.div
                  className="bg-gradient-to-br from-purple-500 to-purple-600 rounded-full p-2 shadow-lg"
                  animate={shouldAnimate ? {
                    rotate: [0, 5, -5, 0],
                  } : {}}
                  transition={{
                    delay: 3.6,
                    duration: 0.8,
                    ease: 'easeInOut',
                  }}
                >
                  <Boxes className="h-7 w-7 text-white" />
                </motion.div>
                <h4 className="text-slate-800 dark:text-slate-100 text-2xl font-bold m-0">
                  Workmachines
                </h4>
              </div>

              {/* Internal components with stagger */}
              <motion.div
                className="grid sm:grid-cols-2 gap-8"
                variants={staggerContainerVariants}
                initial={shouldAnimate ? 'hidden' : 'visible'}
                animate="visible"
              >
                {/* Workspaces Card */}
                <WorkspaceCard shouldAnimate={shouldAnimate} />

                {/* Environments Card */}
                <EnvironmentCard shouldAnimate={shouldAnimate} />
              </motion.div>

              {/* Legend */}
              <motion.div
                className="mt-6 pt-6 border-t border-slate-200 dark:border-slate-700"
                variants={staggerContainerVariants}
                initial={shouldAnimate ? 'hidden' : 'visible'}
                animate="visible"
              >
                <div className="flex flex-wrap gap-4 justify-center text-xs text-slate-600 dark:text-slate-400">
                  <motion.div className="flex items-center gap-2" variants={slideInLeftVariants}>
                    <div className="w-3 h-3 rounded-full bg-gradient-to-r from-blue-500 to-purple-500"></div>
                    <span>Orchestration flow</span>
                  </motion.div>
                  <motion.div className="flex items-center gap-2" variants={slideInLeftVariants}>
                    <div className="w-3 h-3 rounded-full bg-emerald-500"></div>
                    <span>Development</span>
                  </motion.div>
                  <motion.div className="flex items-center gap-2" variants={slideInLeftVariants}>
                    <div className="w-3 h-3 rounded-full bg-amber-500"></div>
                    <span>Infrastructure</span>
                  </motion.div>
                </div>
              </motion.div>
            </motion.div>
          </motion.div>
        </div>
      </div>
    </section>
  )
}

// Workspace Card Component
function WorkspaceCard({ shouldAnimate }: { shouldAnimate: boolean }) {
  return (
    <motion.div className="relative group" variants={fadeInUpVariants}>
      {/* Stack layers */}
      <motion.div
        className="absolute inset-0 bg-gradient-to-br from-emerald-50 to-green-100 dark:from-emerald-950 dark:to-green-900 rounded-xl border-2 border-emerald-200 dark:border-emerald-800 translate-x-2 translate-y-2 opacity-40 group-hover:opacity-60 transition-opacity"
        initial={shouldAnimate ? { opacity: 0, scale: 0.9 } : { opacity: 0.4, scale: 1 }}
        animate={{ opacity: 0.4, scale: 1 }}
        transition={{ delay: 0.1, duration: ANIMATION_DURATION.fast }}
      />
      <motion.div
        className="absolute inset-0 bg-gradient-to-br from-emerald-50 to-green-100 dark:from-emerald-950 dark:to-green-900 rounded-xl border-2 border-emerald-300 dark:border-emerald-700 translate-x-1 translate-y-1 opacity-60 group-hover:opacity-80 transition-opacity"
        initial={shouldAnimate ? { opacity: 0, scale: 0.95 } : { opacity: 0.6, scale: 1 }}
        animate={{ opacity: 0.6, scale: 1 }}
        transition={{ delay: 0.15, duration: ANIMATION_DURATION.fast }}
      />

      {/* Main card */}
      <div className="relative bg-gradient-to-br from-emerald-50 to-green-50 dark:from-emerald-950 dark:to-green-950 rounded-xl border-2 border-emerald-400 dark:border-emerald-600 p-6 hover:shadow-lg transition-all hover:border-emerald-500">
        <div className="flex flex-col items-center gap-3">
          <motion.div
            className="bg-gradient-to-br from-emerald-500 to-green-600 rounded-full p-3 shadow-md"
            animate={shouldAnimate ? {
              rotate: [0, 5, -5, 0],
            } : {}}
            transition={{
              delay: 0.2,
              duration: 0.6,
              ease: 'easeInOut',
            }}
          >
            <Code2 className="h-6 w-6 text-white" />
          </motion.div>
          <h5 className="text-emerald-800 dark:text-emerald-200 text-lg font-bold m-0">
            Workspaces
          </h5>
          <p className="text-emerald-600 dark:text-emerald-400 text-xs text-center m-0">
            Dev Containers
          </p>
        </div>
      </div>
    </motion.div>
  )
}

// Environment Card Component
function EnvironmentCard({ shouldAnimate }: { shouldAnimate: boolean }) {
  return (
    <motion.div className="relative group" variants={fadeInUpVariants}>
      {/* Stack layers */}
      <motion.div
        className="absolute inset-0 bg-gradient-to-br from-amber-50 to-orange-100 dark:from-amber-950 dark:to-orange-900 rounded-xl border-2 border-amber-200 dark:border-amber-800 translate-x-2 translate-y-2 opacity-40 group-hover:opacity-60 transition-opacity"
        initial={shouldAnimate ? { opacity: 0, scale: 0.9 } : { opacity: 0.4, scale: 1 }}
        animate={{ opacity: 0.4, scale: 1 }}
        transition={{ delay: 0.1, duration: ANIMATION_DURATION.fast }}
      />
      <motion.div
        className="absolute inset-0 bg-gradient-to-br from-amber-50 to-orange-100 dark:from-amber-950 dark:to-orange-900 rounded-xl border-2 border-amber-300 dark:border-amber-700 translate-x-1 translate-y-1 opacity-60 group-hover:opacity-80 transition-opacity"
        initial={shouldAnimate ? { opacity: 0, scale: 0.95 } : { opacity: 0.6, scale: 1 }}
        animate={{ opacity: 0.6, scale: 1 }}
        transition={{ delay: 0.15, duration: ANIMATION_DURATION.fast }}
      />

      {/* Main card */}
      <div className="relative bg-gradient-to-br from-amber-50 to-orange-50 dark:from-amber-950 dark:to-orange-950 rounded-xl border-2 border-amber-400 dark:border-amber-600 p-6 hover:shadow-lg transition-all hover:border-amber-500">
        <div className="flex flex-col items-center gap-3">
          <motion.div
            className="bg-gradient-to-br from-amber-500 to-orange-600 rounded-full p-3 shadow-md"
            animate={shouldAnimate ? {
              rotate: [0, 5, -5, 0],
            } : {}}
            transition={{
              delay: 0.2,
              duration: 0.6,
              ease: 'easeInOut',
            }}
          >
            <Database className="h-6 w-6 text-white" />
          </motion.div>
          <h5 className="text-amber-800 dark:text-amber-200 text-lg font-bold m-0">
            Environments
          </h5>
          <p className="text-amber-600 dark:text-amber-400 text-xs text-center m-0">
            Services & Apps
          </p>
        </div>
      </div>
    </motion.div>
  )
}
