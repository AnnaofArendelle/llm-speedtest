import { useEffect } from 'react';
import { animate, motion, useMotionValue, useReducedMotion, useTransform } from 'framer-motion';

/** Animated number that counts from the previous value to the new one. */
export function CountUp({ value, digits = 0, duration = 0.6 }: { value: number; digits?: number; duration?: number }) {
  const reduce = useReducedMotion();
  const mv = useMotionValue(0);
  const text = useTransform(mv, (v) => (digits > 0 ? v.toFixed(digits) : Math.round(v).toLocaleString()));
  useEffect(() => {
    const controls = animate(mv, value, { duration: reduce ? 0 : duration, ease: 'easeOut' });
    return () => controls.stop();
  }, [value, mv, reduce, duration]);
  return <motion.span>{text}</motion.span>;
}
