import { motion } from 'framer-motion';

const BounceIt = ({
  disable = false,
  onClick = (_) => {},
  className = '',
  ...etc
}) => {
  return (
    <motion.div
      className={`${className} inline-block`}
      initial={{ scale: 1 }}
      whileTap={{ scale: disable ? 1 : 0.99 }}
      onClick={onClick}
      {...etc}
    />
  );
};

export default BounceIt;
