import { useEffect } from 'react';

const DataSetter = ({
  set = (_: any) => _,
  value,
}: {
  set: (_: any) => void;
  value: any;
}) => {
  useEffect(() => {
    set(value);
  }, []);
  return null;
};

export default DataSetter;
