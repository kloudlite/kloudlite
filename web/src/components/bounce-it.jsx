import { motion } from 'framer-motion';

export const BounceIt = ({
  disable = false,
  onClick = (_) => {},
  className = '',
  ...etc
}) => {
  if(disable){
    return <div>{etc.children}</div>
  }
  return (
    <motion.div
      tabIndex={"-1"}
      className={`${className} inline-block`}
      initial={{ scale: 1 }}
      whileTap={{ scale: 0.99 }}
      onClick={onClick}
      {...etc}
    />
  );
};
