import AppCompute from '~/iotconsole/page-components/app/compute';
import AppWrapper from '~/iotconsole/page-components/app/app-wrapper';

const SettingCompute = () => {
  return (
    <AppWrapper title="Compute">
      <AppCompute mode="edit" />
    </AppWrapper>
  );
};
export default SettingCompute;
