import { motion } from 'framer-motion';

export const BounceIt = ({
  disable = false,
  onClick = (_) => {},
  className = '',
  ...etc
}) => {
  if(disable){
    return etc.children
  }
  return (
    <motion.div
      tabIndex={"-1"}
      className={`${className} inline-block`}
      initial={{ scale: 1 }}
      whileTap={{ scale: 0.999 }}
      onClick={onClick}
      {...etc}
    />
  );
};
