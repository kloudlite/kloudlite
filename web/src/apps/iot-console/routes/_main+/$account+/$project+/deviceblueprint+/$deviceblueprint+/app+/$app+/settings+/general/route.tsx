import { Box } from '~/iotconsole/components/common-console-components';
import AppWrapper from '~/iotconsole/page-components/app/app-wrapper';
import AppGeneral from '~/iotconsole/page-components/app/general';

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
