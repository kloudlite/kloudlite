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
      initial={{ y: 0 }}
      whileTap={{ y: disable ? 0 : 1 }}
      onClick={onClick}
      {...etc}
    />
  );
};

export default BounceIt;
