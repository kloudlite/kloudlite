import { TitleBox } from '~/console/components/raw-wrapper';
import { ExposedPorts } from '../_.$account.$cluster.$project.$scope.$workspace.new-app/app-network';
import { FadeIn } from '../_.$account.$cluster.$project.$scope.$workspace.new-app/util';

const AppNetwork = () => {
  return (
    <FadeIn>
      <TitleBox
        title="Network"
        subtitle="Expose service ports that need to be exposed from container"
      />
      <ExposedPorts />
    </FadeIn>
  );
};

export default AppNetwork;
