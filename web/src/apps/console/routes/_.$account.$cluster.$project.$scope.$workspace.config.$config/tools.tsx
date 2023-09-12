import CommonTools from '~/console/components/common-tools';
import { IToolsProps } from '~/console/server/utils/common';

const Tools = ({ viewMode, setViewMode }: IToolsProps) => {
  return <CommonTools {...{ viewMode, setViewMode }} options={[]} />;
};

export default Tools;
