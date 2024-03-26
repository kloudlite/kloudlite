import { Box } from '~/console/components/common-console-components';
import AppWrapper from '~/console/page-components/app/app-wrapper';
import AppGeneral from '~/console/page-components/app/general';

const SettingGeneral = () => {
  return (
    <AppWrapper title="General">
      <Box title="Application detail">
        <AppGeneral mode="edit" />
      </Box>
    </AppWrapper>
  );
};
export default SettingGeneral;
