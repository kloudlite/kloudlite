import AppWrapper from '~/iotconsole/page-components/app/app-wrapper';
import { ExposedPorts } from '../../../../new-app/app-network';

const AppNetwork = () => {
  return (
    <AppWrapper title="Network">
      <ExposedPorts />
    </AppWrapper>
  );
};

export default AppNetwork;
